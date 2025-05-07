<template>
  <el-card class="history-panel">
    <div slot="header">
      <span>解析历史</span>
      <el-button
          size="mini"
          @click="clearHistory"
          :disabled="historyList.length === 0"
      >
        清空
      </el-button>
    </div>

    <el-empty v-if="historyList.length === 0" description="暂无历史记录" />

    <div v-else class="history-list">
      <div
          v-for="item in historyList"
          :key="item.id"
          class="history-item"
          @click="handleReParse(item.id)"
      >
        <div class="title">{{ item.title }}</div>
        <div class="meta">
          <span>{{ item.author }}</span>
          <el-divider direction="vertical" />
          <span>{{ item.updatedAt }}</span>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

export default {
  name: 'HistoryPanel',
  computed: {
    ...mapState('history', ['historyList'])
  },
  methods: {
    ...mapMutations('history', ['CLEAR_HISTORY']),
    clearHistory() {
      this.CLEAR_HISTORY()
    },
    handleReParse(id) {
      this.docData = this.historyList.find(item => item.id === id).result
    }
  }
}
</script>

<style scoped>
.history-panel {
  margin-top: 20px;
}

.history-item {
  padding: 10px;
  margin: 5px 0;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.3s;
}

.history-item:hover {
  background: #f5f7fa;
}

.title {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta {
  color: #909399;
  font-size: 0.8em;
  margin-top: 5px;
}
</style>