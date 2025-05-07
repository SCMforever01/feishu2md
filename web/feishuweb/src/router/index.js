import Vue from 'vue'
import Router from 'vue-router'
import HomeView from '../views/Home.vue'
import AuthorizationPage from "@/views/AuthorizationPage";

Vue.use(Router)

export default new Router({
    mode: 'hash',
    routes: [
        { path: '/', name:'Home', component: HomeView },
        {
            path: '/authorization',
            name: 'AuthorizationPage',
            component: AuthorizationPage,
        },
        {
            path: '/login',
            name: 'Login',
            component: () => import('@/views/LoginView.vue'),
        }

    ]
})