import Vue from 'vue'
import Vuex from 'vuex'
import getters from './getters'

Vue.use(Vuex)

const modulesFiles: any = require.context('./modules', true, /\.ts$/)

const modules: any = modulesFiles
  .keys()
  .reduce((modules: any, modulePath: any) => {
    const moduleName = modulePath.replace(/^\.\/(.*)\.\w+$/, '$1')
    const value = modulesFiles(modulePath)
    modules[moduleName] = value.default
    return modules
  }, {})

const store = new Vuex.Store({
  modules,
  getters
})

export default store
