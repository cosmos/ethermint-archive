const Web3 = require('web3')

class Web3pool {
  constructor (providers, selector) {
    this.providers = Web3pool.resolveProviders(providers)
    this.web3s = this.providers.map((p) => new Web3(p))
    this.selector = selector || Web3pool.RR()

    this.web3s.forEach((web3) => {
      console.log('extended web3')
      extendWeb3(web3)
    })
  }

  get web3 () {
    return this.selector(this.web3s)
  }

  /**
   * IPCProvider is not supported
   * For IPC every call should be async, there are some sync requests tests are using
   */
  static resolveProviders (providers) {
    return providers.map((provider) => {
      if (/^http/.test(provider)) {
        return new Web3.providers.HttpProvider(provider)
      }

      return null
    }).filter((p) => p != null)
  }

  // Default selector
  static RR () {
    let i = 0

    return (web3s) => {
      return web3s[i++ % web3s.length]
    }
  }
}

function extendWeb3 (web3) {
  web3._extend({
    property: 'txpool',
    methods: [],
    properties: [
      new web3._extend.Property({
        name: 'status',
        getter: 'txpool_status',
        outputFormatter: function (status) {
          status.pending = web3._extend.utils.toDecimal(status.pending)
          status.queued = web3._extend.utils.toDecimal(status.queued)
          return status
        }
      })
    ]
  })
}

module.exports = Web3pool
