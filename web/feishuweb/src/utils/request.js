import router from '@/router';
import axios from 'axios'

// 创建一个新的axios对象，配置基本的请求参数
const request = axios.create({
    baseURL: 'http://localhost:8080',   // 设置请求的基础URL，即后端接口地址
    timeout: 30000                        // 设置请求超时时间为30秒
})

// request 拦截器
// 这个拦截器可以在请求发送前对请求进行处理
// 比如：可以在请求中统一添加token，或对请求参数进行统一加密等
request.interceptors.request.use(config => {
    const token = localStorage.getItem('token')
    if (token && !config.headers.Authorization) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
}, error => {
    // 如果请求发生错误，打印错误信息以便调试
    console.error('request error: ' + error);
    // 返回Promise.reject以通知请求发生错误
    return Promise.reject(error);
});

// response 拦截器
// 这个拦截器可以在接口响应后对结果进行统一处理
request.interceptors.response.use(
    response => {
        // 获取响应数据
        let res = response.data;

        // 兼容处理服务端返回的字符串数据，将其解析为JSON对象
        if (typeof res === 'string') {
            res = res ? JSON.parse(res) : res;
        }

        // 检查响应中的状态码，如果为401则表示未授权
        if(res.code === '401'){
            // 如果未授权，则重定向到登录页面
            router.push('/login');
        }

        // 返回处理后的响应数据
        return res;
    },
    error => {
        // 如果响应发生错误，打印错误信息以便调试
        console.error('response error: ' + error);
        // 返回Promise.reject以通知响应发生错误
        return Promise.reject(error);
    }
)

// 导出配置好的axios实例，以便在其他模块中使用
export default request;