const fs = require('fs')
const path = require('path')

const tmpdir = '/tmp/truffle-tests'
const name = 'test1'

let ethereum = fs.readFileSync(path.join(tmpdir, name + '.output.ethereum'))
let ethermint = fs.readFileSync(path.join(tmpdir, name + '.output.ethermint'))

ethereum = generateTestObject(ethereum.toString())
ethermint = generateTestObject(ethermint.toString())

diff(ethereum, ethermint)

function generateTestObject (src) {
  let rawdata = '[' + src.toString().substring(0, src.length - 2) + ']'
  let data = JSON.parse(rawdata)

  return data
}

function diff (ethereum, ethermint) {
  if (ethereum.length !== ethermint.length) {
    console.error('Tests count didn\'t match')
    console.error('Ethereum tests: %d, Ethermint tests: %d', ethereum.length, ethermint.length)
    return
  }

  let errors = false
  ethereum.forEach((test, i) => {
    let keys = Object.keys(test)

    keys.forEach((key) => {
      if (ethereum[i][key] !== ethermint[i][key]) {
        errors = true
        console.error('For test case %s\n ethereum(%s) != ethermint(%s)', key, ethereum[i][key], ethermint[i][key])
      }
    })
  })

  if (!errors) {
    console.log('Success: Output for ethereum and ethermint was the same')
  }
}
