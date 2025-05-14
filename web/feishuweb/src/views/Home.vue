<template>
  <div class="home-layout">
    <!-- 渐变背景 -->
    <div class="bg-gradient"></div>
    <!-- 侧边栏 -->
    <div class="sidebar glass" :class="{ 'collapsed': isCollapsed }">
      <div class="sidebar-top">
        <div class="logo-section">
          <img src="@/assets/feishu_lark.png" alt="Logo" class="logo-img" />
          <span class="logo-text">feishuweb</span>
          <div class="collapse-btn" @click="toggleCollapse" :class="{ 'collapsed': isCollapsed }">
            <i :class="['el-icon-d-arrow-left', { 'collapsed': isCollapsed }]"></i>
          </div>
        </div>
        <el-button class="sidebar-btn" type="primary" @click="handleNewParse">
          <img src="@/assets/xinjian.png" alt="新建" class="btn-icon" />
          <span class="btn-text" v-show="!isCollapsed">新建解析</span>
        </el-button>
        <!-- 侧边栏历史记录 -->
        <div class="sidebar-history" v-if="historyList.length">
          <div
              v-for="item in historyList"
              :key="item.id"
              class="sidebar-history-item"
              @click="handleHistoryParse(item.id)"
              :title="item.title"
          >
            <i class="el-icon-document"></i>
            <span class="sidebar-history-title" v-show="!isCollapsed">{{ item.title || item.shortContent }}</span>
          </div>
        </div>
      </div>

      <!-- 用户区域：点击显示菜单框 -->
      <div class="sidebar-bottom" :style="{ width: isCollapsed ? '60px' : '250px' }">
        <div class="user-toggle" @click="toggleUserMenu">
          <img :src="avatarUrl" class="avatar" />
          <div class="user-meta" v-show="!isCollapsed">
            <span class="username">{{ username || '用户' }}</span>
            <i :class="userMenuVisible ? 'el-icon-arrow-down' : 'el-icon-arrow-up'" />
          </div>
        </div>

        <!-- 上拉菜单 -->
        <transition name="fade">
          <div v-show="userMenuVisible" class="user-menu-panel">
            <div class="user-menu-item" @click="handleShowAuthCode">
              <i class="el-icon-key"></i> 查看凭证
            </div>
            <div class="user-menu-item" @click="handleChangePwd">
              <i class="el-icon-lock"></i> 修改密码
            </div>
            <div class="user-menu-item danger" @click="handleLogout">
              <i class="el-icon-switch-button"></i> 退出登录
            </div>
          </div>
        </transition>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-area" :class="{ collapsed: isCollapsed }">
      <div class="header main-title-center" :class="{ collapsed: isCollapsed }">
        飞书文档解析器
      </div>
      <div class="auth-status-bar" v-if="isLogin">
        <el-tag type="success" effect="dark">
          <i class="el-icon-key"></i> 用户凭证
        </el-tag>
      </div>
      <div class="main-content glass">
        <div v-if="isLoading" class="loading">解析中...</div>
        <div v-else-if="docData">
          <div class="header-section">
            <h2 class="doc-title">{{ docData.sheetTitle }}</h2>
            <export-toolbar :content="safeContent" :title="docData.Title" />
          </div>
          <div class="meta-info">
            <el-tag type="info" effect="plain" v-if="docData.author">
              <i class="el-icon-user"></i> {{ docData.author }}
            </el-tag>
            <el-tag type="warning" effect="plain" class="ml-10" v-if="docData.updatedAt">
              <i class="el-icon-time"></i> {{ formattedTime }}
            </el-tag>

          </div>
          <div class="content-box" v-html="safeContent" v-highlight/>
        </div>
        <div v-else class="placeholder">解析数据</div>
      </div>

      <div class="footer-input glass" :class="{ collapsed: isCollapsed }">
        <div class="footer-row">
          <el-input
              v-model="inputUrl"
              placeholder="请输入解析文档链接"
              class="footer-inputbox"
              @keyup.enter.native="handleParse"
          >
            <template #prepend>
              <i class="el-icon-paperclip"></i>
            </template>
          </el-input>
          <el-button type="primary" class="footer-btn right-btn" :loading="isLoading" @click="handleParse">
            立即解析
          </el-button>
        </div>
        <div class="footer-row footer-row-bottom">
          <el-button type="success" class="footer-btn left-btn" @click="handleGetAuthCode">获取授权码</el-button>
          <el-checkbox v-model="withImageDownload" style="margin-left: 20px">下载图片</el-checkbox>
          <input type="file" @change="handleFileChange" accept=".docx,.pdf,.pptx,.xlsx" style="display: inline-block;" />
        </div>
      </div>
    </div>

    <!-- 查看凭证弹窗 -->
    <el-dialog :visible.sync="authDialogVisible" title="凭证信息" width="400px">
      <div class="auth-info">
        <div class="auth-item">
          <div class="auth-label">授权码:</div>
          <div class="auth-value">{{ authCode || '无' }}</div>
          <el-button type="primary" size="small" plain @click="copyToClipboard(authCode)">复制</el-button>
        </div>
        <div class="auth-item">
          <div class="auth-label">访问令牌:</div>
          <div class="auth-value">{{ displayAccessToken }}</div>
          <el-button type="primary" size="small" plain @click="copyToClipboard(accessToken)">复制</el-button>
          <el-button :type="toggleButtonVisible ? 'text' : 'primary'" size="small" @click="toggleAccessTokenDisplay">
            {{ toggleButtonText }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { mapActions, mapMutations, mapState } from 'vuex'
import DOMPurify from 'dompurify'
import ExportToolbar from '@/views/ExportToolbar'
import request from '@/utils/request'
import {marked} from "marked";

export default {
  name: 'HomeView',
  components: {
    ExportToolbar
  },
  data() {
    return {
      inputUrl: '',
      progress: 0,
      interval: null,
      authCode: '',
      showModal: false,
      authUrl: '',
      accessToken: '',
      showHistory: false,
      avatarUrl: require('@/assets/yonghu.png'),
      isLogin: false,
      isCollapsed: false,
      username: '',
      markedContent: '', // 转换后的Markdown内容
      markdownContent: false, // 是否为Markdown格式
      userMenuVisible: false,
      authDialogVisible: false,
      displayAccessToken: '',
      toggleButtonVisible: true,
      withImageDownload: false, // 是否下载图片
      isFile: false, // 是否是文件解析
      file: null, // 上传的文件
      isLoading: false // 定义为data属性
    }
  },
  computed: {
    ...mapState('feishu', ['docData', 'error']),
    ...mapState('history', ['historyList']),

    safeContent() {
      if (!this.docData) return ''
      const rawHtml = marked(this.docData.markdown || '')
      const purifier = DOMPurify.sanitize
      return purifier(rawHtml, {
        ADD_TAGS: ['table', 'thead', 'tbody', 'tr', 'td', 'th'],
        ADD_ATTR: ['colspan', 'rowspan', 'align', 'style']
      })
    },
    formattedTime() {
      return this.docData ? new Date(this.docData.updatedAt).toLocaleString() : ''
    },
    toggleButtonText() {
      return this.toggleButtonVisible ? '显示完整' : '隐藏'
    }
  },
  watch: {
    // 移除了对isLoading的监听，因为现在它是一个data属性
  },
  mounted() {

    if (window.location.search.includes('code')) {
      this.handleRedirect()
    }
    this.isLogin = !!localStorage.getItem('token')
    this.username = localStorage.getItem('user') || ''
    this.authCode = localStorage.getItem('authCode')
    this.accessToken = localStorage.getItem('accessToken')
    this.displayAccessToken = this.formatToken(this.accessToken)
    this.$store.dispatch('history/fetchHistory')
    window.addEventListener('scroll', this.handleHeaderScroll)
  },
  beforeDestroy() {
    window.removeEventListener('scroll', this.handleHeaderScroll)
  },
  methods: {
    ...mapActions('feishu', ['parseDocument']),
    ...mapMutations('history', ['ADD_HISTORY']),
    handleGetAuthCode() {
      if (!this.isLogin) {
        this.$router.push('/login')
        return
      }
      const clientId = 'cli_a72e872fe4fbd00e'
      const redirectUri = 'http://localhost:3000'
      const state = 'RANDOMSTRING'
      const scope =[ "docx:document",
          "docx:document:readonly",
          "drive:drive",
          "drive:drive:readonly",
          "sheets:spreadsheet",
          "sheets:spreadsheet:read",
          "sheets:spreadsheet:readonly",
          "wiki:node:read",
          "wiki:wiki",
          "wiki:wiki:readonly"
          ].join(' ')
      const scopeEncoded = encodeURIComponent(scope)
      this.authUrl = `https://accounts.feishu.cn/open-apis/authen/v1/authorize?client_id=${clientId}&redirect_uri=${redirectUri}&state=${state}&scope=${scopeEncoded}`
      this.showModal = true
      window.location.href = this.authUrl  // 添加此行
    },
    handleRedirect() {
      const urlParams = new URLSearchParams(window.location.search)
      const code = urlParams.get('code')
      const state = urlParams.get('state')

      if (state === 'RANDOMSTRING' && code) {
        this.authCode = code
        this.showModal = false
        localStorage.setItem('authCode', code)
        this.$message.success('授权成功，授权码已获取')
        this.getAccessToken(code)
      } else {
        this.$message.error('授权失败或状态不匹配')
        this.showModal = false
      }
    },
    handleNewParse() {
      // 方法实现
      console.log('handleNewParse clicked')
      // 添加具体逻辑
    },
    convertToMarkdown(htmlContent) {
      try {
        // 使用marked将HTML转换为Markdown
        this.markedContent = marked(htmlContent)
        this.markdownContent = true
      } catch (error) {
        console.error('转换为Markdown格式失败:', error)
        this.markedContent = '转换为Markdown格式失败'
        this.markdownContent = true
      }
    },
    async getAccessToken(code) {
      try {
        const redirectUri = 'http://localhost:3000'
        const requestBody = {
          code: code,
          redirect_uri: redirectUri
        }
        const response = await request.post('/v1/feishu/access_token', requestBody)

        if (response && response.code === 0) {
          let accessToken;
          accessToken = response.content
          this.accessToken = accessToken
          localStorage.setItem('accessToken', accessToken)
          this.displayAccessToken = this.formatToken(accessToken)
          this.$message.success('Token获取成功')
        } else {
          const errorMsg = response?.msg || '未知错误'
          throw new Error(`服务返回错误: ${errorMsg}`)
        }
      } catch (error) {
        console.error('Token请求失败:', error)
        let message = '获取access_token失败'
        if (error.response) {
          switch (error.response.status) {
            case 400:
              message = '请求参数错误'
              break
            case 401:
              message = '认证失败，请检查授权码'
              break
            case 502:
              message = '飞书服务不可用'
              break
            default:
              message = `服务异常: ${error.message || ''}`
              break
          }
        }
        this.$message.error(`${message} (${error.message || ''})`)
        throw error
      }
    },
    handleParse() {
      if (!this.isLogin) {
        this.$router.push('/login')
        return
      }

      if (this.isFile) {
        if (!this.file) {
          this.$message.error('请选择文件')
          return
        }
      } else {
        if (!this.inputUrl) {
          this.$message.error('请输入解析链接')
          return
        }
        if (!this.validateUrl()) return
      }

      let userId
      const userInfo = localStorage.getItem('userInfo')
      if (!userInfo) {
        this.$message.error('用户信息不存在，请重新登录')
        this.$router.push('/login')
        return
      }

      try {
        const parsedUserInfo = JSON.parse(userInfo)
        userId = parsedUserInfo.id
      } catch (e) {
        console.error('解析用户信息失败:', e)
        this.$message.error('解析用户信息失败，请重新登录')
        localStorage.removeItem('userInfo')
        this.$router.push('/login')
        return
      }

      const user_access_token = localStorage.getItem('accessToken')
      if (!user_access_token) {
        this.$message.error('未找到访问令牌或者已经过期，请重新登录')
        this.clearUserSession()
        return
      }

      this.isLoading = true
      try {
        this.parseDocument({
          url: this.inputUrl,
          id: String(userId),
          user_access_token: user_access_token,
          withImageDownload: this.withImageDownload,
          isFile: this.isFile,
          file: this.file // 如果是文件解析，传递文件对象
        }).then(response => {
          if (!response || typeof response.code === 'undefined') {
            //this.$message.error('服务器返回格式错误')
            return
          }
              if (response.code === 0) {
                this.$message.success('解析成功')
                this.safeContent = response.data.markdown  // 添加这行
                if (this.docData) {
                  this.ADD_HISTORY({
                    ...this.docData,
                    url: this.inputUrl
                  })
                  this.convertToMarkdown(this.docData.markdown)
                }
              } else {
                this.$message.error(`解析失败: ${response.message || '未知错误'}`)
              }
            })
            .catch(error => {
              console.error('解析错误:', error)
              this.$message.error('解析过程中发生错误')
            })
            .finally(() => {
              this.isLoading = false
            })
      } catch (error) {
        console.error('解析错误:', error)
        this.$message.error('解析过程中发生错误')
        this.isLoading = false
      }
    },
    clearUserSession() {
      localStorage.clear()
      this.$store.commit('SET_TOKEN', '')
      this.$store.commit('SET_USER_INFO', null)
      this.isLogin = false
      this.username = ''
      this.$router.push('/login')
      this.$message.success('已退出登录')
    },

    validateUrl() {
      const pattern = /^https:\/\/[a-zA-Z0-9-]+\.feishu\.cn\/(docs|wiki)\/[a-zA-Z0-9]+(\?.*)?/;
      if (!pattern.test(this.inputUrl)) {
        this.$message.error('请输入有效的飞书文档链接')
        return false
      }
      return true
    },
    handleHistoryParse(id) {
      this.docData = this.historyList.find(item => item.id === id).result
    },
    startProgress() {
      this.progress = 0
      this.interval = setInterval(() => {
        if (this.progress < 90) {
          this.progress += 10
        }
      }, 300)
    },
    clearProgress() {
      clearInterval(this.interval)
      this.interval = null
      this.progress = 0
    },
    handleUserCenter() {
      if (!this.isLogin) {
        this.$router.push('/login')
      } else {
        this.$router.push('/user-center')
      }
    },
    handleLogout() {
      this.$confirm('确认退出登录吗？', '退出登录', {
        confirmButtonText: '退出',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        localStorage.clear()
        this.$store.commit('SET_TOKEN', '')
        this.$store.commit('SET_USER_INFO', null)
        this.isLogin = false
        this.username = ''
        this.$router.push('/login')
        this.$message.success('已退出登录')
      }).catch(() => {})
    },
    handleChangePwd() {
      this.$prompt('请输入新密码', '修改密码', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'password'
      }).then(({ value }) => {
        this.$message.success(`密码已更新（模拟）: ${value}`)
      }).catch(() => {})
    },
    handleShowAuthCode() {
      this.displayAccessToken = this.formatToken(this.accessToken)
      this.toggleButtonVisible = true
      this.authDialogVisible = true
    },
    toggleCollapse() {
      this.isCollapsed = !this.isCollapsed
    },
    toggleUserMenu() {
      this.userMenuVisible = !this.userMenuVisible
    },
    handleHeaderScroll() {
      const scrollY = window.scrollY || window.pageYOffset
      const max = 120
      const percent = Math.min(scrollY / max, 1)

      const bgAlpha = 0.4 + 0.4 * percent
      const shadowAlpha = 0.0 + 0.12 * percent
      const translateY = -percent * 24
      const blurValue = 4 + 8 * percent

      document.documentElement.style.setProperty('--header-bg', `rgba(255,255,255,${bgAlpha})`)
      document.documentElement.style.setProperty('--header-shadow', `0 2px 16px rgba(0,0,0,${shadowAlpha})`)
      document.documentElement.style.setProperty('--header-transform', `translate(-50%, ${translateY}px)`)
      document.documentElement.style.setProperty('--header-opacity', 0.92 + 0.08 * percent)
      document.documentElement.style.setProperty('--header-blur-filter', `blur(${blurValue}px)`)
    },
    formatToken(token) {
      if (!token || token === '无') return token
      const start = token.substring(0, 6)
      const end = token.substring(token.length - 6)
      return `${start}.......${end}`
    },
    copyToClipboard(text) {
      if (!text || text === '无') {
        this.$message.warning('无内容可复制')
        return
      }
      navigator.clipboard.writeText(text).then(() => {
        this.$message.success('已复制到剪贴板')
      }, (err) => {
        this.$message.error('复制失败')
        console.error('Could not copy text: ', err)
      })
    },
    toggleAccessTokenDisplay() {
      if (this.toggleButtonVisible) {
        this.displayAccessToken = this.accessToken
      } else {
        this.displayAccessToken = this.formatToken(this.accessToken)
      }
      this.toggleButtonVisible = !this.toggleButtonVisible
    },
    handleFileChange(e) {
      this.file = e.target.files[0]
      this.isFile = true
    }
  }
}
</script>

<style scoped>
.content-box table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1em;
}

.content-box th,
.content-box td {
  border: 1px solid #ccc;
  padding: 8px 12px;
  text-align: left;
}

.content-box th {
  background: #f0f0f0;
}

.home-layout {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: row;
  overflow: hidden;
}

.bg-gradient {
  position: fixed;
  z-index: 0;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%, #fcb69f 100%, #ffecd2 100%);
  background-size: 400% 400%;
  animation: gradientMove 15s ease infinite;
}
@keyframes gradientMove {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}
.sidebar {
  width: 250px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: center;
  padding: 0;
  z-index: 2;
  box-shadow: 2px 0 16px 0 rgba(0,0,0,0.04);
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar.collapsed {
  width: 60px;
}

.logo-section {
  position: relative;
  width: 100%;
  padding: 20px 16px;
  display: flex;
  align-items: center;
  margin-bottom: 32px;
  justify-content: flex-start;
  height: 60px;
}

.logo-img {
  height: 50px;
  width: auto;
  transition: all 0.3s ease;
  filter: brightness(0.2);
}

.logo-text {
  font-size: 25px;
  font-weight: 600;
  color: #333;
  margin-left: 15px;
  transition: opacity 0.3s ease;
}

.collapsed .logo-text {
  display: none;
}

.collapsed .logo-img {
  height: 34px;
  margin: 0 auto;
}

.collapse-btn {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 6px;
  background: rgba(79, 77, 77, 0.04);
  transition: all 0.3s ease;
  position: absolute;
}

.collapse-btn:hover {
  background: rgba(79, 77, 77, 0.04);
}

.collapse-btn i {
  transition: transform 0.1s ease;
  color: #666;
  font-size: 14px;
}

.collapse-btn i.collapsed {
  transform: rotate(180deg);
}

.collapse-btn {
  right: 18px;
  top: 50%;
  transform: translateY(-50%);
}

.sidebar.collapsed .collapse-btn {
  right: 50%;
  bottom: -32px;
  top: auto;
  transform: translateX(50%);
}

.sidebar.collapsed .logo-section {
  justify-content: center;
  flex-direction: column;
  align-items: center;
  padding-left: 0;
  padding-right: 0;
}
.sidebar.collapsed .logo-img {
  margin: 0 auto !important;
  display: block;
}

.sidebar-top {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 0;
}

.sidebar-btn {
  margin: 0 24px !important;
  width: calc(100% - 48px) !important;
  background: rgba(64, 158, 255, 0.08) !important;
  color: #409EFF !important;
  border: none !important;
  border-radius: 12px;
  box-shadow: none !important;
  transition: all 0.3s;
}

.sidebar-btn:hover {
  background: rgba(64, 158, 255, 0.15) !important;
}

.sidebar-history {
  width: 100%;
  padding: 0 12px 12px 12px;
  margin-top: 20px;
  flex: 1 1 auto;
  overflow-y: auto;
}
.sidebar-history-item {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  color: #333;
  font-size: 15px;
  margin-bottom: 4px;
  transition: background 0.3s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sidebar-history-item:hover {
  background: #f5f7fa;
}
.sidebar-history-item i {
  margin-right: 8px;
  color: #409EFF;
  font-size: 18px;
}
.sidebar-history-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sidebar.collapsed .sidebar-history-title {
  display: none;
}

.sidebar-btn .btn-icon {
  width: 26px;
  height: 22px;
  margin-right: 12px;
}

.collapsed .sidebar-btn {
  width: 44px !important;
  height: 44px !important;
}

.sidebar-btn .btn-text {
  font-size: 18px;
}

.sidebar-bottom {
  position: relative;
  padding: 10px;
}

.user-toggle {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: background 0.2s;
}
.user-toggle:hover {
  background: #f5f5f5;
}
.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
}
.user-meta {
  display: flex;
  align-items: center;
  margin-left: 10px;
  font-size: 14px;
  font-weight: 500;
  gap: 6px;
}
.user-menu-panel {
  margin-top: 8px;
  padding: 10px;
  background: #fdfdfd;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}
.user-menu-item {
  display: flex;
  align-items: center;
  font-size: 14px;
  padding: 8px;
  cursor: pointer;
  border-radius: 6px;
}
.user-menu-item i {
  margin-right: 6px;
}
.user-menu-item:hover {
  background: #f0f0f0;
}
.user-menu-item.danger {
  color: #f56c6c;
}
.user-menu-item.danger:hover {
  background: #fef0f0;
}
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
  min-width: 0;
  margin-left: 220px;
  padding: 0 32px;
  margin-bottom: 110px;
  height: 100vh;
  overflow: auto;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-top: 80px;
}

.main-area.collapsed {
  margin-left: 90px;
}

.footer-input {
  position: fixed;
  left: 290px;
  right: 90px;
  bottom: 32px;
  transition: left 0.3s cubic-bezier(0.4, 0,0.2, 1);
  height: 120px;
  padding-top: 0;
  padding-bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.footer-input.collapsed {
  left: 180px;
}

.header {
  position: fixed;
  top: 0;
  left: 50%;
  transform: var(--header-transform, translate(-50%, 0));
  width: 800px;
  z-index: 100;
  background: var(--header-bg, rgba(255, 255, 255, 0.4));
  box-shadow: var(--header-shadow, 0 2px 16px rgba(0, 0, 0, 0));
  opacity: var(--header-opacity, 0.92);
  backdrop-filter: var(--header-blur-filter, blur(4px));
  -webkit-backdrop-filter: var(--header-blur-filter, blur(4px));
  transition: all 0.3s ease;
  text-align: center;
  font-size: 2.2em;
  font-weight: 550;
  color: #757373;
  border-radius: 8px;
  letter-spacing: 2px;
  text-shadow: 0 2px 8px rgba(255, 255, 255, 0.2);
  padding: 24px 0 18px 0;
}

.header.collapsed {
  left: 275px;
  width: 1000px;
}
.main-title-center {
  text-align: center;
}
.auth-status-bar {
  position: absolute;
  top: 32px;
  right: 40px;
  display: flex;
  gap: 12px;
  z-index: 10;
}
.main-content {
  flex: 1;
  margin: 0;
  margin-bottom: 90px;
  border-radius: 12px;
  min-height: 400px;
  padding: 32px;
  position: relative;
  display: flex;
  flex-direction: column;
  background: rgba(255,255,255,0.8);
}
.placeholder {
  font-size: 2.2em;
  color: #bdbdbd;
  text-align: center;
  margin-top: 120px;
  letter-spacing: 4px;
}
.footer-input {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 12px 16px;
  border-radius: 30px;
  box-shadow: 0 8px 32px 0 rgba(31,38,135,0.07);
  background: rgba(255,255,255,0.7);
  z-index: 20;
  backdrop-filter: blur(16px);
  border: 2px solid #e3e6ef;
  overflow: visible;
  min-height: 60px;
}
.footer-row {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 8px;
}
.footer-row-bottom {
  margin-top: 8px;
  justify-content: flex-start;
}
.footer-btn {
  min-width: 120px;
  font-size: 1em;
  border-radius: 10px;
  padding: 8px 0;
  height: 36px;
  background: rgba(64, 158, 255, 0.08) !important;
  border: none !important;
  color: #409EFF !important;
}
.footer-btn:hover {
  background: rgba(64, 158, 255, 0.15) !important;
  transform: translateY(-1px);
}
.footer-btn.right-btn {
  min-width: 120px;
  border-radius: 15px;
}
.footer-btn.left-btn {
  min-width: 140px;
  border-radius: 15px;
  background: rgba(103, 194, 58, 0.08) !important;
  color: #67C23A !important;
}
.footer-btn.left-btn:hover {
  background: rgba(103, 194, 58, 0.15) !important;
}
.footer-inputbox {
  flex: 1;
  border-radius: 20px;
  background: #f7faff;
}

.footer-input ::v-deep(.el-input__inner) {
  height: 36px;
  line-height: 36px;
}

.footer-input ::v-deep(.el-input-group__prepend),
.footer-input ::v-deep(.el-input-group__append) {
  padding: 0 12px;
  height: 36px;
  line-height: 36px;
}
.glass {
  background: rgba(255,255,255,0.7) !important;
  backdrop-filter: blur(10px) !important;
  box-shadow: 0 8px 32px 0 rgba(31,38,135,0.07);
  border: 1px solid rgba(255, 255, 255, 0.3);
}
.loading {
  font-size: 1.5em;
  color: #888;
  text-align: center;
  margin-top: 120px;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter, .fade-leave-to {
  opacity: 0;
}

.auth-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.auth-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.auth-label {
  font-weight: bold;
  min-width: 80px;
}
.auth-value {
  flex: 1;
  word-break: break-all;
}
</style>