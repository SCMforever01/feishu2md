export default function setupMock() {
    if (process.env.NODE_ENV === 'development') {
        // 直接加载mock文件，避免声明未使用的变量
       // require('mockjs')
        //require('../mocks/feishu')
    }
}