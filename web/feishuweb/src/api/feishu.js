import request from '../utils/request'

export const parseFeishuDoc = (data, token) => {
    return request({
        url: '/v1/transform',
        method: 'post',
        headers: {
            Authorization: `Bearer ${token}`
        },
        data: data
    })
}