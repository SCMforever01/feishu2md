<template>
  <div class="export-toolbar">
    <el-dropdown @command="handleExport">
      <el-button type="primary">
        导出文档<i class="el-icon-arrow-down el-icon--right"></i>
      </el-button>
      <el-dropdown-menu slot="dropdown">
        <el-dropdown-item command="pdf">PDF格式</el-dropdown-item>
        <el-dropdown-item command="md">Markdown格式</el-dropdown-item>
        <el-dropdown-item command="html">HTML格式</el-dropdown-item>
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
import jsPDF from 'jspdf'
import TurndownService from 'turndown'

export default {
  name: 'ExportToolbar',
  props: {
    content: { type: String, required: true },
    title: { type: String, required: true }
  },
  methods: {
    async copyMarkdown() {
      const turndownService = new TurndownService()
      const markdown = turndownService.turndown(this.content)
      try {
        await navigator.clipboard.writeText(markdown)
        this.$message.success('已复制到剪贴板')
      } catch (err) {
        console.error('复制失败:', err)
        this.$message.error('复制失败，请手动复制')
      }
    },

    handleExport(type) {
      switch (type) {
        case 'pdf': this.exportPDF(); break
        case 'md': this.exportMarkdown(); break
        case 'html': this.exportHTML(); break
      }
    },
    exportPDF() {
      const doc = new jsPDF()
      const turndownService = new TurndownService()
      const markdown = turndownService.turndown(this.content)
      const lines = markdown.split('\n')
      lines.forEach((line, index) => {
        doc.text(line, 10, 10 + index * 10)
      })
      doc.save(`${this.sanitizeFilename(this.title)}.pdf`)
    },
    exportMarkdown() {
      this.downloadFile(this.content, 'text/markdown', 'md')
    },
    exportHTML() {
      this.downloadFile(this.content, 'text/html', 'html')
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
