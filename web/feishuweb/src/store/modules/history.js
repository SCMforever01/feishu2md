import request from '@/utils/request'

const state = {
    historyList: []
}

const mutations = {
    SET_HISTORY(state, list) {
        state.historyList = list
    },
    CLEAR_HISTORY(state) {
        state.historyList = []
    }
}

const actions = {
    async fetchHistory({ commit }) {
        try {
            const token = localStorage.getItem('token') // 从 localStorage 中读取 JWT

            if (!token) return

            const res = await request.get('/v1/getHistory', {
                headers: {
                    Authorization: `Bearer ${token}`
                }
            })

            if (res.code === 200) {
                commit('SET_HISTORY', res.data.map(item => ({
                    ...item,
                    shortContent: item.result.slice(0, 80).replace(/\n/g, ' ') + '...'
                })))
            }
        } catch (err) {
            console.error('Failed to fetch history:', err)
        }
    }
}

export default {
    namespaced: true,
    state,
    mutations,
    actions
}
