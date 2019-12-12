import { API_URL, PRECISION } from '../../common/config'
import axios from 'axios'
import router from '../../router'
import { Response, Parser } from 'rivine-ts-types'
import store from '../index'

const parser: Parser = new Parser(PRECISION)

const explorer = {
  state: {
    explorer: Object,
    block: Object,
    hash:  Response,
    loading: true,
    success: true,
    error: String,
    darkMode: true
  },
  mutations: {
    SET_EXPLORER: (state: any, explorer: object) => {
      state.explorer = explorer
    },
    SET_BLOCK: (state: any, block: object) => {
      state.block = block
    },
    SET_HASH: (state: any, hash: Response) => {
      state.hash = hash
    },
    SET_LOADING: (state: any, loading: boolean) => {
      state.loading = loading
    },
    SET_SUCCESS: (state: any, success: boolean) => {
      state.success = success
    },
    SET_ERROR: (state: any, error: string) => {
      state.error = error
    },
    SET_DARK_MODE: (state: any, darkMode: string) => {
      state.darkMode = darkMode
    }
  },
  actions: {
    SET_EXPLORER: async (context: any) => {
      await axios({ method: 'GET', url: API_URL + '/explorer' }).then(
        result => {
          context.commit('SET_EXPLORER', result.data)
        },
        error => {
          if (error.response.status === 400) {
            router.push('/notfound')
          }
        }
      )
    },
    SET_BLOCK_HEIGHT: async ({ dispatch, commit }, height: Number) => {
      // tslint:disable-next-line
      axios({
        method: 'GET',
        url: API_URL + '/explorer/blocks/' + height
      }).then(
        result => {
          const parsedResponse = parser.ParseBlockResponseJSON(result.data)
          commit('SET_BLOCK', parsedResponse)
        },
        error => {
          if (error.response.status === 400) {
            commit('SET_ERROR', `${height}`)
            router.push('/notfound')
          }
        }
      )
    },
    SET_HASH: async (context: any, hash: string) => {
      if (!hash) return
      let parsedResponse
      await context.commit('SET_LOADING', true)
      await axios({
        method: 'GET',
        url: API_URL + '/explorer/hashes/' + hash
      }).then(
        result => {
          if (result.data === '') {
            context.commit('SET_ERROR', `${hash}`)
            router.push('/notfound')
          }
          switch (result.data.hashtype) {
            case 'blockid':
              parsedResponse = parser.ParseBlockResponseJSON(result.data)
              context.commit('SET_BLOCK', parsedResponse)
              context.commit('SET_LOADING', false)
              // route to block page
              router.push(`/block/${parsedResponse.height}`)
              break
            default:
              parsedResponse = parser.ParseHashResponseJSON(result.data, hash)
              context.commit('SET_HASH', parsedResponse)
              context.commit('SET_LOADING', false)
          }
        },
        error => {
          if (error.response.status === 400) {
            context.commit('SET_ERROR', `${hash}`)
            router.push('/notfound')
          }
        }
      )
    },
    SET_ERROR: (context: any, error: string) => {
      context.commit('SET_ERROR', error)
    },
    SET_DARK_MODE: (context: any, darkMode: boolean) => {
      context.commit('SET_DARK_MODE', darkMode)
    },
    SET_LOADING: (context: any, loading: boolean) => {
      context.commit('SET_LOADING', loading)
    }
  }
}

export default explorer
