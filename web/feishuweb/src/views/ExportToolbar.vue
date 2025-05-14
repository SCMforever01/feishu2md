<template>
  <div class="export-toolbar">
    <el-dropdown @command="handleExport">
      <el-button type="primary">
        导出文档 <i class="el-icon-arrow-down el-icon--right"></i>
      </el-button>
      <el-dropdown-menu slot="dropdown">
        <el-dropdown-item command="pdf">PDF格式</el-dropdown-item>
        <el-dropdown-item command="md">Markdown格式（含图片）</el-dropdown-item>
        <el-dropdown-item command="html">HTML格式（含图片）</el-dropdown-item>
      </el-dropdown-menu>
    </el-dropdown>

    <el-button
        type="success"
        icon="el-icon-document-copy"
        size="mini"
        style="margin-left: 12px"
        @click="copyMarkdown"
    >
      复制 Markdown
    </el-button>
  </div>
</template>

<script>
import TurndownService from 'turndown'
import html2pdf from 'html2pdf.js'

export default {
  name: 'ExportToolbar',
  props: {
    content: { type: String, required: true }, // HTML 字符串，包含图片标签
    title: { type: String, required: true }    // 导出文件标题
  },
  methods: {
    async handleExport(type) {
      switch (type) {
        case 'pdf': await this.exportPDF(); break
        case 'md': await this.exportMarkdown(); break
        case 'html': await this.exportHTML(); break
      }
    },

    async exportMarkdown() {
      console.log('🚀 当前 this.content 内容:\n', this.content)
      const htmlWithBase64 = await this.replaceImgWithBase64(this.content)
      const turndownService = new TurndownService()
      const markdown = turndownService.turndown(htmlWithBase64)
      this.downloadFile(markdown, 'text/markdown', 'md')
    },

    async exportHTML() {
      const htmlWithBase64 = await this.replaceImgWithBase64(this.content)
      this.downloadFile(htmlWithBase64, 'text/html', 'html')
    },

    async exportPDF() {
      const htmlWithBase64 = await this.replaceImgWithBase64(this.content)
      const container = document.createElement('div')
      container.innerHTML = htmlWithBase64

      const opt = {
        margin: 0.5,
        filename: `${this.sanitizeFilename(this.title)}.pdf`,
        image: { type: 'jpeg', quality: 0.98 },
        html2canvas: { scale: 2 },
        jsPDF: { unit: 'in', format: 'a4', orientation: 'portrait' }
      }

      html2pdf().set(opt).from(container).save()
    },

    async copyMarkdown() {
      const htmlWithBase64 = await this.replaceImgWithBase64(this.content)
      const turndownService = new TurndownService()
      const markdown = turndownService.turndown(htmlWithBase64)

      try {
        await navigator.clipboard.writeText(markdown)
        this.$message.success('已复制到剪贴板')
      } catch (err) {
        console.error('复制失败:', err)
        this.$message.error('复制失败，请手动复制')
      }
    },

    async replaceImgWithBase64(html) {
      const div = document.createElement('div')
      div.innerHTML = html
      const imgTags = div.querySelectorAll('img')

      for (const img of imgTags) {
        try {
          const base64 = await this.convertImageToBase64(img.src)
          img.src = base64
        } catch (e) {
          console.warn('图片转 base64 失败:', img.src, e)
        }
      }

      return div.innerHTML
    },

    async convertImageToBase64(url) {
      const response = await fetch(url)
      const blob = await response.blob()
      return new Promise(resolve => {
        const reader = new FileReader()
        reader.onloadend = () => resolve(reader.result)
        reader.readAsDataURL(blob)
      })
    },

    downloadFile(content, mimeType, extension) {
      const blob = new Blob([content], { type: mimeType })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = `${this.sanitizeFilename(this.title)}.${extension}`
      link.click()
      URL.revokeObjectURL(link.href)
    },

    sanitizeFilename(name) {
      return name.replace(/[\\/:*?"<>|]/g, '_')
    }
  }
}
</script>

<style scoped>
.export-toolbar {
  margin-top: 16px;
  text-align: right;
}
</style>
