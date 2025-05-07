import Vue from 'vue'
import Vuex from 'vuex'
import feishu from './modules/feishu'
import history from './modules/history'
Vue.use(Vuex)

export default new Vuex.Store({
    modules: {
        feishu,
        history
    }
})