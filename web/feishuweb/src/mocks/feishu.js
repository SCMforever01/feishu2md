// const Mock = require('mockjs')
// Mock.mock(/\/v1\/transform/, 'post', (options) => {
//     const url = JSON.parse(options.body).url
//     return Mock.mock({
//         code: 200,
//         data: {
//             id: '@guid',
//             title: '@ctitle(15, 30)',
//             author: '@cname',
//             updatedAt: '@datetime',
//             content: Mock.Random.text(500, 1000),
//             type: /table/.test(url) ? 'table' : 'text'
//         }
//     })
// })

