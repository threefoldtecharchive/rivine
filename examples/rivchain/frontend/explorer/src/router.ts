import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'
import Block from './views/Block.vue'
import NotFound from './views/NotFound.vue'
import Hash from './views/Hash.vue'
import Charts from './views/Charts.vue'
import transactions from './views/Transactions.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/transactions/',
      name: 'transactions',
      component: transactions
    },
    {
      path: '/block/:block?',
      name: 'block',
      component: Block
    },
    {
      path: '/hashes/:hash?',
      name: 'hash',
      component: Hash
    },
    {
      path: '/charts',
      name: 'charts',
      component: Charts
    },
    {
      path: '/notfound',
      component: NotFound
    },
    {
      path: '*',
      component: NotFound
    }
  ]
})
