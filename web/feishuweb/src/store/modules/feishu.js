// store/feishu.js
import { parseFeishuDoc } from '@/api/feishu'

const state = () => ({
    isLoading: false,
    docData: null,
    error: null
})

const mutations = {
    SET_LOADING(state, status) {
        state.isLoading = status
    },
    SET_DOC_DATA(state, data) {
        state.docData = data
    },
    SET_ERROR(state, error) {
        state.error = error
    }
}

const actions = {
    async parseDocument({ commit }, { url, id, user_access_token, withImageDownload, isFile }) {
        commit('SET_LOADING', true)
        commit('SET_ERROR', null)
        try {
            const token = localStorage.getItem('token')
            if (!token) {
                commit('SET_ERROR', '未登录')
                return
            }

            const data = {
                id: id,
                url: url,
                collection: 'default', // 默认集合，可根据需求修改
                access_key: token,
                user_access_token: user_access_token,
                with_image_download: withImageDownload,
                is_file: isFile
            }

            const response = await parseFeishuDoc(data, token)
            commit('SET_DOC_DATA', response.data)
        } catch (error) {
            commit('SET_ERROR', error.message || '解析失败')
        } finally {
            commit('SET_LOADING', false)
        }
    }
}

export default {
    namespaced: true,
    state,
    mutations,
    actions
}