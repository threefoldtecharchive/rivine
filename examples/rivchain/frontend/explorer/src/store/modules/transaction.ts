import { API_URL } from '../../common/config'
import axios from 'axios'

const transactions = {
  state: {
    transactions: Array
  },
  mutations: {
    SET_TRANSACTIONS: (state: any, transactions: Array<Object>) => {
      state.transactions = transactions
    }
  },
  actions: {
    SET_TRANSACTIONS: async (context: any) => {
      await axios({
        method: 'GET',
        url: API_URL + '/transactionpool/transactions'
      }).then(result => {
        context.commit('SET_TRANSACTIONS', result.data)
      })
    }
  }
}

export default transactions
