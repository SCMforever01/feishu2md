package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"feishu2md/server/internal/config"
	"feishu2md/server/internal/model"
	"fmt"
	"github.com/Wsine/feishu2md/core"
	"github.com/chyroc/lark"
	"log"
	"math/rand"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	client *lark.Lark
}

func NewClient(appID, appSecret, domain string) *Client {
	return &Client{
		client: lark.New(
			lark.WithAppCredential(appID, appSecret),
			lark.WithOpenBaseURL("https://open."+domain),
			lark.WithTimeout(60*time.Second),
		),
	}
}

func (c *Client) GetWikiNodeInfo(ctx context.Context, token string, UserAccessToken string) (*lark.GetWikiNodeRespNode, error) {

	var resp *lark.GetWikiNodeResp
	var err error

	// 判断 UserAccessToken 是否为空
	if UserAccessToken != "" {
		// 如果 UserAccessToken 不为空，使用 WithUserAccessToken 作为选项
		resp, _, err = c.client.Drive.GetWikiNode(ctx, &lark.GetWikiNodeReq{
			Token: token,
		}, lark.WithUserAccessToken(UserAccessToken))
	} else {
		// 如果 UserAccessToken 为空，不使用任何额外选项
		resp, _, err = c.client.Drive.GetWikiNode(ctx, &lark.GetWikiNodeReq{
			Token: token,
		})
	}
	if err != nil {
		return nil, err
	}
	return resp.Node, nil
}

func (c *Client) DownloadFile(fileToken, userAccessToken string) (*lark.DownloadDriveFileResp, error) {
	resp, _, err := c.client.Drive.DownloadDriveFile(
		context.Background(),
		&lark.DownloadDriveFileReq{
			FileToken: fileToken, // 注意确认 SDK 实际要求的字段名称
		},
		lark.WithUserAccessToken(userAccessToken),
	)
	return resp, err
}

// GetDocumentContent 获取文档内容
func (c *Client) GetDocumentContent(ctx context.Context, docToken, userAccessToken string) (*model.DocContentResult, error) { //获取真实文件数据
	// 1. 获取基础文档内容
	docx, blocks, tittle, err := c.GetDocxContent(ctx, docToken, userAccessToken)
	if err != nil {
		return nil, fmt.Errorf("获取文档内容失败: %w", err)
	}
	// 2. 空文档检查
	if len(blocks) == 0 {
		return nil, fmt.Errorf("文档内容为空")
	}

	// 3. 构建块索引映射
	indexMap := make(map[string]int)
	for i, block := range blocks {
		indexMap[block.BlockID] = i
	}
	// 4. 重组文档结构
	for _, block := range blocks {
		// 处理文本类型和标题类型的块
		if isTextOrHeading(block) && hasChildren(block) {
			moveChildrenToRoot(block, blocks, indexMap)
			block.Children = []string{} // 清空原块子节点
		}

		// 处理列表类型块
		if isListBlock(block) {
			filterCodeChildren(block, blocks, indexMap)
		}
	}
	// 5. 排序根块子节点
	sortChildrenByIndex(blocks[0].Children, indexMap)

	// 6. 处理表格块
	for n := 0; n < len(blocks); n++ {
		block := blocks[n]
		if block == nil {
			continue
		}

		if isTableBlock(block) {
			newBlocks, err := c.processTableBlock(ctx, block, userAccessToken, docToken)
			if err != nil {
				return nil, err
			}

			// 插入新生成的块
			blocks = insertBlocks(blocks, n+1, newBlocks)
			n += len(newBlocks) // 跳过新增块

			// 更新根块子节点
			updateRootChildren(blocks[0], block.BlockID, newBlocks[0].BlockID)
		}
	}

	// 7. 转换为Markdown
	markdown, imgTokens := parseDocxContent(docx, blocks)

	// 8. 获取文档标题（假设从某个地方提取标题）
	docTitle := tittle // 这里可以根据文档结构获取标题，或是从其它源提取

	// 返回文档内容和标题
	return &model.DocContentResult{
		Markdown:  markdown,
		DocTitle:  docTitle,
		ImgTokens: imgTokens,
	}, nil
}

func parseDocxContent(docx *lark.DocxDocument, blocks []*lark.DocxBlock) (string, []string) {
	cfg := config.LoadConfig()
	newConfig := core.NewConfig(cfg.Feishu.AppID, cfg.Feishu.AppSecret)
	parser := core.NewParser(newConfig.Output)
	return parser.ParseDocxContent(docx, blocks), parser.ImgTokens
}

// GetDocxContent 获取普通文档内容
func (c *Client) GetDocxContent(ctx context.Context, docToken, userAccessToken string) (*lark.DocxDocument, []*lark.DocxBlock, string, error) {
	// 创建请求
	req := &lark.GetDocxDocumentReq{
		DocumentID: docToken,
	}
	var resp *lark.GetDocxDocumentResp
	var err error

	// 根据是否提供用户访问令牌进行请求
	if userAccessToken != "" {
		resp, _, err = c.client.Drive.GetDocxDocument(ctx, req, lark.WithUserAccessToken(userAccessToken))
	} else {
		resp, _, err = c.client.Drive.GetDocxDocument(ctx, req)
	}

	if err != nil {
		return nil, nil, "", err
	}

	docx := &lark.DocxDocument{
		DocumentID: resp.Document.DocumentID,
		RevisionID: resp.Document.RevisionID,
		Title:      resp.Document.Title,
	}

	var blocks []*lark.DocxBlock
	var pageToken *string

	for {
		// 获取文档的块列表
		blockReq := &lark.GetDocxBlockListOfDocumentReq{
			DocumentID: docx.DocumentID,
			PageToken:  pageToken,
		}

		if userAccessToken != "" {
			resp2, _, err := c.client.Drive.GetDocxBlockListOfDocument(ctx, blockReq, lark.WithUserAccessToken(userAccessToken))
			if err != nil {
				return docx, nil, "", err
			}
			blocks = append(blocks, resp2.Items...)
			pageToken = &resp2.PageToken
			if !resp2.HasMore {
				break
			}
		} else {
			resp2, _, err := c.client.Drive.GetDocxBlockListOfDocument(ctx, blockReq)
			if err != nil {
				return docx, nil, "", err
			}
			blocks = append(blocks, resp2.Items...)
			pageToken = &resp2.PageToken
			if !resp2.HasMore {
				break
			}
		}
	}

	return docx, blocks, resp.Document.Title, nil
}

func (c *Client) DownloadImageRaw(ctx context.Context, imgToken, imgDir string, userAccessToken string) (string, []byte, error) {

	var resp *lark.DownloadDriveMediaResp
	var err error

	// 判断 userAccessToken 是否为空
	if userAccessToken != "" {
		// 调用 DownloadDriveMedia 方法并传递选项
		resp, _, err = c.client.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
			FileToken: imgToken,
		}, lark.WithUserAccessToken(userAccessToken))
	} else {
		// 调用 DownloadDriveMedia 方法不传递选项
		resp, _, err = c.client.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
			FileToken: imgToken,
		})
	}
	if err != nil {
		return imgToken, nil, err
	}
	fileext := filepath.Ext(resp.Filename)
	filename := fmt.Sprintf("%s/%s%s", imgDir, imgToken, fileext)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.File)
	return filename, buf.Bytes(), nil
}

// ---------- 以下是辅助函数 ----------

func isTextOrHeading(block *lark.DocxBlock) bool {
	return block.BlockType == lark.DocxBlockTypeText ||
		(block.BlockType >= lark.DocxBlockTypeHeading1 && block.BlockType <= lark.DocxBlockTypeHeading9)
}

func hasChildren(block *lark.DocxBlock) bool {
	return len(block.Children) > 0
}

func moveChildrenToRoot(block *lark.DocxBlock, blocks []*lark.DocxBlock, indexMap map[string]int) {
	for _, childID := range block.Children {
		if childIndex, exists := indexMap[childID]; exists {
			blocks[childIndex].ParentID = blocks[0].BlockID
			blocks[0].Children = append(blocks[0].Children, childID)
		}
	}
}

func isListBlock(block *lark.DocxBlock) bool {
	return block.BlockType == lark.DocxBlockTypeOrdered || block.BlockType == lark.DocxBlockTypeBullet
}

func filterCodeChildren(block *lark.DocxBlock, blocks []*lark.DocxBlock, indexMap map[string]int) {
	newChildren := make([]string, 0)
	for _, childID := range block.Children {
		if blocks[indexMap[childID]].BlockType == lark.DocxBlockTypeCode {
			blocks[indexMap[childID]].ParentID = blocks[0].BlockID
			blocks[0].Children = append(blocks[0].Children, childID)
		} else {
			newChildren = append(newChildren, childID)
		}
	}
	block.Children = newChildren
}

func sortChildrenByIndex(children []string, indexMap map[string]int) {
	sort.Slice(children, func(i, j int) bool {
		return indexMap[children[i]] < indexMap[children[j]]
	})
}

func isTableBlock(block *lark.DocxBlock) bool {
	return block.BlockType == lark.DocxBlockTypeSheet || block.BlockType == lark.DocxBlockTypeBitable
}

func (c *Client) processTableBlock(ctx context.Context, block *lark.DocxBlock, userToken, docToken string) ([]*lark.DocxBlock, error) {
	var (
		i, j       int64
		flatValues []string
		merges     []*lark.GetSheetRespSheetMerge
		err        error
	)

	switch block.BlockType {
	case lark.DocxBlockTypeSheet:
		i, j, flatValues, merges, err = c.GetSheetContent(ctx, block.Sheet.Token, userToken)
	case lark.DocxBlockTypeBitable:
		i, j, flatValues, err = c.GetBitableContent(ctx, block.Bitable.Token, userToken)
	}

	if err != nil {
		return nil, fmt.Errorf("获取表格内容失败: %w", err)
	}

	return c.CreateTable(i, j, flatValues, merges, docToken)
}

func insertBlocks(blocks []*lark.DocxBlock, pos int, newBlocks []*lark.DocxBlock) []*lark.DocxBlock {
	return append(blocks[:pos], append(newBlocks, blocks[pos:]...)...)
}

func updateRootChildren(rootBlock *lark.DocxBlock, originalID, newID string) {
	for i, id := range rootBlock.Children {
		if id == originalID {
			rootBlock.Children = append(
				rootBlock.Children[:i+1],
				append([]string{newID}, rootBlock.Children[i+1:]...)...)
			break
		}
	}
}

func (c *Client) GetSheetContent(ctx context.Context, docToken string, userAccessToken string) (int64, int64, []string, []*lark.GetSheetRespSheetMerge, error) {
	var err error

	// 解析文档令牌
	parts := strings.Split(docToken, "_")
	if len(parts) != 2 {
		return 0, 0, nil, nil, fmt.Errorf("invalid docToken format")
	}
	sheetToken, sheetID := parts[0], parts[1]
	// 1. 获取合并信息
	sheetReq1 := &lark.GetSheetReq{
		SpreadSheetToken: sheetToken,
		SheetID:          sheetID,
	}
	fmt.Printf("sheetReq1:%v", sheetReq1)
	var sheetResp1 *lark.GetSheetResp
	if userAccessToken != "" {
		sheetResp1, _, err = c.client.Drive.GetSheet(ctx, sheetReq1, lark.WithUserAccessToken(userAccessToken))
	} else {
		sheetResp1, _, err = c.client.Drive.GetSheet(ctx, sheetReq1)
	}
	if sheetResp1 != nil {
		jsonStr, _ := json.MarshalIndent(sheetResp1, "", "  ")
		fmt.Printf("sheetResp1 JSON:\n%s\n", jsonStr)
	}
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("获取表格元数据失败: %w", err)
	}
	merges := sheetResp1.Sheet.Merges

	// 2. 获取范围数据
	valueRenderOpt := "UnformattedValue"
	dateTimeRenderOpt := "FormattedString"
	batchReq := &lark.BatchGetSheetValueReq{
		SpreadSheetToken:     sheetToken, // 路径参数
		Ranges:               []string{fmt.Sprintf("%s", sheetID)},
		ValueRenderOption:    &valueRenderOpt, // 获取原始值
		DateTimeRenderOption: &dateTimeRenderOpt,
	}
	var batchResp *lark.BatchGetSheetValueResp
	if userAccessToken != "" {
		batchResp, _, err = c.client.Drive.BatchGetSheetValue(ctx, batchReq, lark.WithUserAccessToken(userAccessToken))
	} else {
		batchResp, _, err = c.client.Drive.BatchGetSheetValue(ctx, batchReq)
	}
	if batchResp != nil {
		jsonStr, _ := json.MarshalIndent(batchResp, "", "  ")
		fmt.Printf("batchResp JSON:\n%s\n", jsonStr)
	}
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("批量获取表格数据失败: %w", err)
	}
	// 处理并打印结果
	rowCount, colCount, flatValues, err := processValues(batchResp, merges)
	flatJson, err := json.MarshalIndent(flatValues, "", "  ")
	if err != nil {
		fmt.Println("flatValues 转换 JSON 出错:", err)
	} else {
		fmt.Printf("flatValues JSON:\n%s\n", flatJson)
	}

	if err != nil {
		fmt.Println("处理数据时出错: %v", err)
	}
	return rowCount, colCount, flatValues, merges, err
}

func (c *Client) GetBitableContent(ctx context.Context, bitableToken string, userAccessToken string) (int64, int64, []string, error) {
	parts := strings.Split(bitableToken, "_")
	var viewID string
	if len(parts) == 3 {
		viewID = parts[2]
	}
	req := &lark.GetBitableRecordListReq{
		AppToken: parts[0],
		TableID:  parts[1],
		ViewID:   &viewID,
	}

	var resp *lark.GetBitableRecordListResp
	var err error
	if userAccessToken != "" {
		resp, _, err = c.client.Bitable.GetBitableRecordList(ctx, req, lark.WithUserAccessToken(userAccessToken))
	} else {
		resp, _, err = c.client.Bitable.GetBitableRecordList(ctx, req)
	}

	if err != nil {
		return 0, 0, nil, err
	}

	if resp == nil || len(resp.Items) == 0 {
		return 0, 0, nil, fmt.Errorf("no data available in response")
	}

	// 行数
	rowCount := resp.Total + 1

	// 查找最大字段数和字段名
	maxFields := 0
	var fieldNames []string
	for _, item := range resp.Items {
		if item == nil || item.Fields == nil {
			continue
		}
		numFields := len(item.Fields)
		if numFields > maxFields {
			maxFields = numFields
			fieldNames = make([]string, 0, numFields)
			for k := range item.Fields {
				fieldNames = append(fieldNames, k)
			}
		}
	}

	colCount := int64(maxFields)
	if colCount == 0 {
		return rowCount, colCount, nil, fmt.Errorf("no fields found in any item")
	}

	// 创建扁平化数组
	flatValues := make([]string, rowCount*colCount)

	// 填充表头
	copy(flatValues, fieldNames)

	// 填充数据
	for itemIdx, item := range resp.Items {
		if item == nil || item.Fields == nil {
			continue
		}
		for fieldIdx, field := range fieldNames {
			idx := (itemIdx+1)*int(colCount) + fieldIdx
			flatValues[idx] = parseFieldValue(item.Fields, field)
		}
	}

	return rowCount, colCount, flatValues, nil
}

func (c *Client) CreateTable(i int64, j int64, flatValues []string, merges []*lark.GetSheetRespSheetMerge, docToken string) ([]*lark.DocxBlock, error) {
	// 存储最终的块数组
	var blocksArray []*lark.DocxBlock
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
	//生成tableID
	tableID := randomString(24)
	// 初始化 columnWidth 并分配默认宽度
	columnWidth := make([]int64, j)
	for idx := range columnWidth {
		columnWidth[idx] = 100 // 默认宽度
	}

	// 初始化 cellChildren 并生成随机字符串填充
	cellChildren := make([]string, 0, i*j)
	for k := 0; k < int(i)*int(j); k++ {
		cellChildren = append(cellChildren, randomString(24)) // 生成长度为24的随机字符串
	}
	mergeInfo := make([]*lark.DocxBlockTablePropertyMergeInfo, 0, i*j)
	for k := 0; k < int(i)*int(j); k++ {
		// 创建新的 mergeInfo 对象，设置 col_span 和 row_span 为 1
		info := &lark.DocxBlockTablePropertyMergeInfo{
			ColSpan: 1,
			RowSpan: 1,
		}
		// 将对象追加到 mergeInfo 列表中
		mergeInfo = append(mergeInfo, info)
	}
	// 如果 merges 不为空，则处理合并区域
	if len(merges) > 0 {
		// 遍历 merges 数据，更新 mergeInfo 中的 RowSpan 和 ColSpan
		for _, merge := range merges {
			startRow := merge.StartRowIndex
			endRow := merge.EndRowIndex
			startCol := merge.StartColumnIndex
			endCol := merge.EndColumnIndex

			// 计算 RowSpan 和 ColSpan
			rowSpan := endRow - startRow + 1
			colSpan := endCol - startCol + 1

			// 更新 mergeInfo 数组中对应的单元格
			for row := startRow; row <= endRow; row++ {
				for col := startCol; col <= endCol; col++ {
					index := int(row)*int(j) + int(col)
					if row == startRow && col == startCol {
						// 对合并区域左上角的单元格设置 RowSpan 和 ColSpan
						mergeInfo[index].RowSpan = rowSpan
						mergeInfo[index].ColSpan = colSpan
					} else {
						// 非左上角的合并单元格，将 RowSpan 和 ColSpan 设为 0
						mergeInfo[index].RowSpan = 0
						mergeInfo[index].ColSpan = 0
					}
				}
			}
		}
	}

	tableProperty := &lark.DocxBlockTableProperty{
		ColumnSize:  j,
		ColumnWidth: columnWidth,
		MergeInfo:   mergeInfo,
		RowSize:     i,
	}

	table := &lark.DocxBlockTable{
		Cells:    cellChildren,
		Property: tableProperty,
	}
	tableBlock := &lark.DocxBlock{
		ParentID:  docToken,
		BlockType: 31,
		BlockID:   tableID,
		Children:  cellChildren,
		Table:     table,
	}
	// 将 tableBlock 加入 blocksArray
	blocksArray = append(blocksArray, tableBlock)

	textChildren := make([]string, 0, i*j)
	for k := 0; k < int(i)*int(j); k++ {
		textChildren = append(textChildren, randomString(24)) // 生成长度为24的随机字符串
	}
	for n, childID := range cellChildren {

		cellBlock := &lark.DocxBlock{
			BlockID:   childID,
			ParentID:  tableID,
			BlockType: 32,
			Children:  []string{textChildren[n]},
		}
		if len(flatValues) <= n {
			return nil, fmt.Errorf("flatValues length is insufficient")
		}
		textRun := &lark.DocxTextElementTextRun{
			Content: flatValues[n],
		}
		textElement := &lark.DocxTextElement{
			TextRun: textRun,
		}
		style := &lark.DocxTextStyle{
			Align:  1,     // 对齐方式为居左
			Folded: false, // 折叠状态为 false
		}
		blockText := &lark.DocxBlockText{
			Style:    style,
			Elements: []*lark.DocxTextElement{textElement},
		}
		textblock := &lark.DocxBlock{
			BlockID:   textChildren[n],
			ParentID:  childID,
			BlockType: 2,
			Text:      blockText,
		}
		// 将 cellBlock 和 textblock 按顺序加入 blocksArray
		blocksArray = append(blocksArray, cellBlock, textblock)
	}

	return blocksArray, nil
}

func (c *Client) GetSheetInfo(ctx context.Context, SheetToken string, userAccessToken string) (string, error) {
	var resp *lark.GetSheetListResp
	var err error

	// 创建请求对象
	req := &lark.GetSheetListReq{
		SpreadSheetToken: SheetToken,
	}

	// 检查用户访问令牌并调用获取表格元数据的方法
	if userAccessToken != "" {
		resp, _, err = c.client.Drive.GetSheetList(ctx, req, lark.WithUserAccessToken(userAccessToken))
	} else {
		resp, _, err = c.client.Drive.GetSheetList(ctx, req)
	}

	// 检查错误并返回
	if err != nil {
		return "", fmt.Errorf("failed to get sheet metadata: %w", err)
	}

	// 检查返回的表格是否为空
	if len(resp.Sheets) == 0 {
		return "", fmt.Errorf("no sheets found in the response")
	}

	return resp.Sheets[0].SheetID, nil
}

func (c *Client) GetSheetsContent(ctx context.Context, token string, userAccessToken string, Url string) (*model.SheetContentResult, error) {
	var (
		blocks      []*lark.DocxBlock
		err         error
		sheetID     string
		sheetToken  string
		i, j        int64
		flatValues  []string
		sheetTitle  string
		titleBlock  *lark.DocxBlock
		blocksArray []*lark.DocxBlock
		wg          sync.WaitGroup
		merges      []*lark.GetSheetRespSheetMerge
	)
	// 创建一个带缓冲的错误通道
	errCh := make(chan error, 4)
	// 使用sync.Once确保通道只关闭一次
	var onceSheetToken sync.Once
	var onceContentReady sync.Once

	// 中间通道，用于顺序控制
	sheetTokenCh := make(chan struct{})   // 用于控制sheetToken生成后
	contentReadyCh := make(chan struct{}) // 用于控制sheet内容生成后

	// 1. 获取sheetID
	wg.Add(1)
	go func() {
		defer wg.Done()
		parsedUrl, err := url.Parse(Url)
		if err != nil {
			errCh <- fmt.Errorf("failed to parse URL: %w", err)
			return
		}
		sheetID = parsedUrl.Query().Get("sheet")

		// 如果没有提供 sheetID，使用 token 获取
		if sheetID == "" {
			sheetID, err = c.GetSheetInfo(ctx, token, userAccessToken)
			if err != nil {
				errCh <- fmt.Errorf("failed to get sheet ID: %w", err)
				return
			}
		}
		fmt.Printf("SheetID:%v", sheetID)
		sheetToken = token + "_" + sheetID
		onceSheetToken.Do(func() { close(sheetTokenCh) }) // 生成sheetToken后发出信号
	}()
	// 2. 获取sheet内容，需要等待sheetToken完成
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sheetTokenCh // 等待sheetToken生成信号
		i, j, flatValues, merges, err = c.GetSheetContent(ctx, sheetToken, userAccessToken)
		if err != nil {
			errCh <- err
			return
		}
		onceContentReady.Do(func() { close(contentReadyCh) }) // 生成sheet内容后发出信号
	}()

	// 3. 获取sheet标题，可以并行
	wg.Add(1)
	go func() {
		defer wg.Done()
		sheetTitle, err = c.GetSheetTitle(ctx, token, userAccessToken)
		if err != nil {
			errCh <- err
			return
		}
		titleBlock, err = c.CreateTitleBlock(sheetTitle)
		if err != nil {
			errCh <- err
		}
	}()
	// 4. 创建表格，依赖于sheet内容的生成
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-contentReadyCh // 等待sheet内容生成信号
		blocksArray, err = c.CreateTable(i, j, flatValues, merges, token)
		if err != nil {
			errCh <- err
			return
		}
	}()

	// 启动一个goroutine来等待所有任务完成并关闭错误通道
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// 检查是否有错误
	for err := range errCh {
		if err != nil {
			fmt.Println("Error:", err)
			return &model.SheetContentResult{
				Markdown:   "",
				SheetTitle: "",
			}, err
		}
	}
	titleBlock.Children = append(titleBlock.Children, blocksArray[0].BlockID)
	blocksArray[0].ParentID = titleBlock.BlockID
	// 创建标题块

	blocks = append(blocks, titleBlock)
	blocks = append(blocks, blocksArray...)
	if err != nil {
		return &model.SheetContentResult{
			Markdown:   "",
			SheetTitle: "",
		}, err
	}
	docx := &lark.DocxDocument{
		DocumentID: titleBlock.BlockID,
		Title:      sheetTitle,
	}
	markdown, imgTokens := parseDocxContent(docx, blocks)
	return &model.SheetContentResult{
		Markdown:   markdown,
		SheetTitle: sheetTitle,
		ImgTokens:  imgTokens,
	}, nil
}

func (c *Client) GetSheetTitle(ctx context.Context, SheetToken string, userAccessToken string) (string, error) {
	var resp *lark.GetSpreadsheetResp
	var err error

	// 创建请求对象
	req := &lark.GetSpreadsheetReq{
		SpreadSheetToken: SheetToken,
	}

	// 检查用户访问令牌并调用获取表格信息的方法
	if userAccessToken != "" {
		resp, _, err = c.client.Drive.GetSpreadsheet(ctx, req, lark.WithUserAccessToken(userAccessToken))
	} else {
		resp, _, err = c.client.Drive.GetSpreadsheet(ctx, req)
	}

	// 检查错误
	if err != nil {
		return "", fmt.Errorf("failed to get spreadsheet title: %w", err)
	}

	// 返回标题或默认值
	if resp.Spreadsheet.Title == "" {
		return "未命名表格", nil
	}
	return resp.Spreadsheet.Title, nil
}

func (c *Client) CreateTitleBlock(title string) (*lark.DocxBlock, error) {

	//生成tableID
	titleBlockID := randomString(24)
	textRun := &lark.DocxTextElementTextRun{
		Content: title,
	}
	textElement := &lark.DocxTextElement{
		TextRun: textRun,
	}
	style := &lark.DocxTextStyle{
		Align:  1,     // 对齐方式为居左
		Folded: false, // 折叠状态为 false
	}
	page := &lark.DocxBlockText{
		Style:    style,
		Elements: []*lark.DocxTextElement{textElement},
	}
	children := make([]string, 0)
	titleBlock := &lark.DocxBlock{
		BlockID:   titleBlockID,
		BlockType: 1,
		Page:      page,
		Children:  children,
	}
	return titleBlock, nil
}

func (c *Client) GetBitablesContent(ctx context.Context, token string, userAccessToken string, Url string) (string, error) {
	var (
		err          error
		i, j         int64
		bitableToken string
		bitableName  string
		flatValues   []string
		titleBlock   *lark.DocxBlock
		blocksArray  []*lark.DocxBlock
		merges       []*lark.GetSheetRespSheetMerge
	)
	// 解析URL参数
	parsedUrl, err := url.Parse(Url)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}
	// 获取表格参数
	tableID := parsedUrl.Query().Get("table")
	viewID := parsedUrl.Query().Get("view")
	// 获取默认表格ID（如果URL未指定）
	if tableID == "" {
		tableID, err = c.GetBitableTableID(ctx, token, userAccessToken)
		if err != nil {
			return "", fmt.Errorf("failed to get default table ID: %w", err)
		}
	}
	// 获取表格元信息
	bitableName, err = c.GetBitableMetainfo(ctx, token, userAccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to get base info: %w", err)
	}
	// 处理视图信息
	bitableToken = token + "_" + tableID
	if viewID != "" {
		viewName, err := c.GetBitableViewName(ctx, bitableToken+"_"+viewID, userAccessToken)
		if err != nil {
			return "", fmt.Errorf("failed to get view name: %w", err)
		}
		bitableName += "_" + viewName
		bitableToken += "_" + viewID
	}
	// 获取表格内容
	i, j, flatValues, err = c.GetBitableContent(ctx, bitableToken, userAccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to get table content: %w", err)
	}

	// 创建标题块
	titleBlock, err = c.CreateTitleBlock(bitableName)
	if err != nil {
		return "", fmt.Errorf("failed to create title block: %w", err)
	}

	// 生成表格块
	blocksArray, err = c.CreateTable(i, j, flatValues, merges, token)
	if err != nil {
		return "", fmt.Errorf("failed to create table blocks: %w", err)
	}

	// 构建块层级关系
	titleBlock.Children = append(titleBlock.Children, blocksArray[0].BlockID)
	blocksArray[0].ParentID = titleBlock.BlockID

	// 组合所有块
	allBlocks := append([]*lark.DocxBlock{titleBlock}, blocksArray...)

	// 创建文档结构
	docx := &lark.DocxDocument{
		DocumentID: titleBlock.BlockID,
		Title:      bitableName,
	}
	markdown, _ := parseDocxContent(docx, allBlocks)
	return markdown, nil
}

func (c *Client) GetBitableTableID(ctx context.Context, appToken string, userAccessToken string) (string, error) {
	var (
		pageToken *string
		tableID   string
	)
	for {
		// 创建请求对象
		req := &lark.GetBitableTableListReq{
			AppToken:  appToken,
			PageToken: pageToken,
			PageSize:  new(int64), // 设置分页大小为 1，取第一个表格即可
		}
		*req.PageSize = 1

		// 调用 API
		resp, _, err := c.client.Bitable.GetBitableTableList(ctx, req, lark.WithUserAccessToken(userAccessToken))
		if err != nil {
			return "", fmt.Errorf("failed to fetch table list: %w", err)
		}

		// 检查返回结果
		if len(resp.Items) > 0 {
			tableID = resp.Items[0].TableID
			break
		}
		if !resp.HasMore {
			break
		}
		pageToken = &resp.PageToken // 更新分页标记
	}

	// 如果未找到任何表格，返回错误
	if tableID == "" {
		return "", fmt.Errorf("no table found for app token %s", appToken)
	}

	return tableID, nil
}

func (c *Client) GetBitableMetainfo(ctx context.Context, BitableToken string, userAccessToken string) (string, error) {
	var resp *lark.GetBitableMetaResp
	var err error

	// 创建请求对象
	req := &lark.GetBitableMetaReq{
		AppToken: BitableToken,
	}

	// 检查用户访问令牌并调用获取表格元数据的方法
	if userAccessToken != "" {
		resp, _, err = c.client.Bitable.GetBitableMeta(ctx, req, lark.WithUserAccessToken(userAccessToken))
	} else {
		resp, _, err = c.client.Bitable.GetBitableMeta(ctx, req)
	}

	// 检查错误并返回
	if err != nil {
		return "", fmt.Errorf("failed to get sheet metadata: %w", err)
	}

	// 检查返回的表格是否为空
	if resp.App == nil {
		return "", fmt.Errorf("no sheets found in the response")
	}
	if resp.App.Name == "" {
		return "", fmt.Errorf("no sheets found in the response")
	}
	return resp.App.Name, nil
}

func (c *Client) GetBitableViewName(ctx context.Context, bitableToken string, userAccessToken string) (string, error) {
	// 检查 bitableToken 格式是否正确
	parts := strings.Split(bitableToken, "_")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid bitableToken format: expected AppToken_TableID_ViewID, got %s", bitableToken)
	}

	// 提取 AppToken, TableID, ViewID
	appToken := parts[0]
	tableID := parts[1]
	viewID := parts[2]

	// 构建请求对象
	req := &lark.GetBitableViewReq{
		AppToken: appToken,
		TableID:  tableID,
		ViewID:   viewID,
	}

	// 发起请求
	resp, _, err := c.client.Bitable.GetBitableView(ctx, req, lark.WithUserAccessToken(userAccessToken))
	if err != nil {
		return "", fmt.Errorf("failed to fetch bitable view: %w", err)
	}

	// 检查响应结果
	if resp == nil || resp.View == nil {
		return "", fmt.Errorf("no view data found for token: %s", bitableToken)
	}

	// 返回视图名称
	return resp.View.ViewName, nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 生成随机字符串
func randomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// 获取GetBitableContent的辅助函数
func parseFieldValue(fields map[string]interface{}, field string) string {
	value, exists := fields[field]
	if !exists {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		return extractNameFromSlice(v)
	case float64:
		return formatFloat64ToDate(v)
	default:
		log.Printf("Field %s has an unexpected type: %T\n", field, v)
		return ""
	}
}

func extractNameFromSlice(v []interface{}) string {
	if len(v) > 0 {
		if firstElem, ok := v[0].(map[string]interface{}); ok {
			if name, ok := firstElem["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}

const dateFormat = "2006-01-02"

func formatFloat64ToDate(value float64) string {
	var t time.Time
	if value > 1e12 {
		t = time.UnixMilli(int64(value))
	} else {
		t = time.Unix(int64(value), 0)
	}
	return t.Format(dateFormat)
}

// 解析sheet数据并处理
func processValues(apiResponse *lark.BatchGetSheetValueResp, merges []*lark.GetSheetRespSheetMerge) (int64, int64, []string, error) {
	var flatValues []string
	var totalRows, totalCols int64
	// 确定最大行列
	for _, valueRange := range apiResponse.ValueRanges {
		values := valueRange.Values
		if len(values) == 0 {
			continue
		}
		rowCount := int64(len(values))
		colCount := int64(len(values[0]))
		totalRows += rowCount
		if colCount > totalCols {
			totalCols = colCount
		}
	}
	for _, merge := range merges {
		if merge.EndColumnIndex+1 > totalCols {
			totalCols = merge.EndColumnIndex + 1
		}
		if merge.EndRowIndex+1 > totalRows {
			totalRows = merge.EndRowIndex + 1
		}
	}
	// 构建表格
	table := make([][]interface{}, totalRows)
	for i := range table {
		table[i] = make([]interface{}, totalCols)
	}
	// 填充值
	for _, valueRange := range apiResponse.ValueRanges {
		values := valueRange.Values
		for rowIndex, row := range values {
			for colIndex, value := range row {
				if rowIndex < len(table) && colIndex < len(table[rowIndex]) {
					table[rowIndex][colIndex] = value
				}
			}
		}
	}
	// 填充合并区域为 nil
	for _, merge := range merges {
		for row := merge.StartRowIndex; row <= merge.EndRowIndex; row++ {
			for col := merge.StartColumnIndex; col <= merge.EndColumnIndex; col++ {
				if row > merge.StartRowIndex || col > merge.StartColumnIndex {
					table[row][col] = nil
				}
			}
		}
	}
	// 扁平化并格式化输出
	for row := 0; row < int(totalRows); row++ {
		for col := 0; col < int(totalCols); col++ {
			cell := table[row][col]
			if cell == nil {
				flatValues = append(flatValues, "")
			} else if sc, ok := cell.(*lark.SheetContent); ok {
				flatValues = append(flatValues, formatSheetContent(sc))
			} else if sc, ok := cell.(lark.SheetContent); ok {
				flatValues = append(flatValues, formatSheetContent(&sc))
			} else {
				flatValues = append(flatValues, fmt.Sprintf("%v", cell))
			}
		}
	}

	return totalRows, totalCols, flatValues, nil
}
func formatSheetContent(cell *lark.SheetContent) string {
	if cell == nil {
		return ""
	}
	if cell.String != nil {
		return *cell.String
	}
	if cell.Int != nil {
		return strconv.FormatInt(*cell.Int, 10)
	}
	if cell.Float != nil {
		return strconv.FormatFloat(*cell.Float, 'f', -1, 64)
	}
	if cell.Formula != nil {
		return "[Formula]"
	}
	if cell.Link != nil {
		return "[Link]"
	}
	if cell.AtUser != nil {
		return "[@User]"
	}
	if cell.AtDoc != nil {
		return "[@Doc]"
	}
	if cell.MultiValue != nil {
		return "[MultiValue]"
	}
	if cell.EmbedImage != nil {
		return "[Image]"
	}
	if cell.Attachment != nil {
		return "[Attachment]"
	}
	if cell.Children != nil {
		return "[Children]"
	}
	return ""
}
