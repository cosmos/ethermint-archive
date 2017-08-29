'use strict'

if (!process.env.WEB3_HOST) {
    throw new Error('Please define WEB3_HOST env variable')
}
if (!process.env.WEB3_PORT) {
    throw new Error('Please define WEB3_PORT env variable')
}

module.exports = {
    host: process.env.WEB3_HOST,
    port: process.env.WEB3_PORT,
}