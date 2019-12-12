<template lang="html">
  <div>
    <navigation />
    <!-- <div class="container" v-if="noRecentTransaction">
      <h1>No unconfirmed Transactions</h1>
      <hr/>
    </div> -->
    <v-content>
      <v-container
        class="fill-height"
        fluid
      >
        <div class="container">
          <h1>Recent Blocks</h1>
          <div
            v-for="(block, idx) in flatten(recentBlockTransactions)"
            v-bind:key="idx"
          >
            <v-simple-table>
              <thead>
                <tr>
                  <th colspan="3">Block: {{ block.height }}</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>Timestamp</td>
                  <td>
                    {{ block.timestamp }}
                  </td>
                </tr>
                <tr v-for="(tx, index) in block.txs" v-bind:key="index">
                  <td>#{{ index + 1 }} Transaction ID</td>
                  <td class="clickable" v-on:click="routeToHashPage(tx.id)">
                    {{ tx.id }}
                  </td>
                </tr>
              </tbody>
            </v-simple-table>
            <br />
          </div>
        </div>
      </v-container>
    </v-content>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import Navigation from '../components/Common/Navigation.vue'
import axios from 'axios'
import { API_URL } from '../common/config'
import { flatten } from 'lodash'

@Component({
  name: 'Transactions',
  components: {
    Navigation
  },
  watch: {
    '$store.state.block': function () {
      this.$router.push('/block/' + this.$store.state.block.block.height)
    }
  },
  methods: {
    routeToHashPage: function (val) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    },
    routeToBlockPage: function (val) {
      this.$store.dispatch('SET_HASH', val)
    }
  }
})
export default class Transactions extends Vue {
  recentBlockTransactions: any = []
  noRecentTransaction: boolean = true

  created () {
    window.scrollTo(0, 0)
    this.$store.dispatch('SET_TRANSACTIONS').then(() => {
      this.$store.dispatch('SET_EXPLORER').then(() => {
        const explorerHeight = this.$store.getters.EXPLORER.height
        for (let i = explorerHeight; i > explorerHeight - 20; i--) {
          axios({ method: 'GET', url: API_URL + '/explorer/blocks/' + i }).then(
            result => {
              this.recentBlockTransactions.push({
                txs: result.data.block.transactions,
                timestamp: this.formatBlockDate(
                  result.data.block.rawblock.timestamp
                ),
                height: result.data.block.height
              })
            }
          )
        }
      })
      if (!this.$store.getters.TRANSACTIONS) {
        this.noRecentTransaction = false
      }
    })
    const _this = this
    setInterval(function () {
      _this.$store.dispatch('SET_TRANSACTIONS').then(() => {
        if (!_this.$store.getters.TRANSACTIONS) {
          _this.noRecentTransaction = false
        }
      })
    }, 60000)
  }

  flatten (some: any) {
    return flatten(some)
  }

  formatBlockDate (timestamp: number) {
    const blockDate = new Date(timestamp * 1000)
    const day = blockDate.getDate()
    const month = blockDate.toLocaleString('default', { month: 'long' })
    const year = blockDate.getFullYear()
    const hours = blockDate.getHours()
    const tempMinutes = blockDate.getMinutes()
    const minutes = tempMinutes < 10 ? `0${tempMinutes}` : tempMinutes

    return `${hours}:${minutes}, ${month} ${day}, ${year}`
  }
}
</script>
<style scoped>
.spinner {
  margin: "auto";
  margin-top: 50px;
  height: 500px;
}
</style>
