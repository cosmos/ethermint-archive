const mocha = require('mocha')
const fs = require('fs')

module.exports = JsonFileReporter

function JsonFileReporter (runner) {
  mocha.reporters.Base.call(this, runner)

  const fileToAppend = process.env.TESTS_FILE
  let obj = {}

  runner.on('pass', (test) => {
    obj[formatName(test)] = ''
    process.stdout.write('.')
  })

  runner.on('fail', (test, err) => {
    obj[formatName(test)] = err.message
    process.stdout.write('-')
  })

  runner.on('end', () => {
    fs.appendFileSync(fileToAppend, JSON.stringify(obj))
  })
}

function formatName (test) {
  return (test.file ? test.file + ':' : '') + test.fullTitle()
}
