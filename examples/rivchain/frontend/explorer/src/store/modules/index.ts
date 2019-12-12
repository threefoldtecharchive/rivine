import Vue from 'vue'
import Vuex from 'vuex'
import camelCase from 'lodash/camelCase'

Vue.use(Vuex)

const requireModule = require.context('.', false, /\.ts$/)
const modules: any = {}

requireModule.keys().forEach(fileName => {
  if (fileName === './index.ts') return

  const moduleName: string = camelCase(fileName.replace(/(\.\/|\.ts)/g, ''))

  modules[moduleName] = requireModule(fileName).default
})

export default modules
