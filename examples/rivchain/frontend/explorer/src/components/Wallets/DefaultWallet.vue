<template lang="html">
  <div>
    <h1>Hash</h1>
    <div>
      <v-simple-table >
        <thead>
          <tr>
            <th colspan="3">Wallet Address</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Address</td>
            <td>{{ wallet.address }}</td>
          </tr>
          <tr>
            <td>Confirmed Coin Balance</td>
            <td>{{ renderValue(wallet.confirmedCoinBalance) }}</td>
          </tr>
          <tr v-if="wallet.lastCoinSpent">
            <td>Last Coin Spend</td>
            <td>
              @ Block:
              <span
                class="clickable"
                v-on:click="routeToBlockPage(wallet.lastCoinSpent.height)"
                >{{ wallet.lastCoinSpent.height }}</span
              >
              Txid:
              <span
                class="clickable"
                v-on:click="routeToHashPage(wallet.lastCoinSpent.txid)"
                >{{ wallet.lastCoinSpent.txid }}</span
              >
            </td>
          </tr>
          <tr>
            <td>Confirmed Block Stake Balance</td>
            <td>{{ wallet.confirmedBlockstakeBalance }} BS</td>
          </tr>
          <tr v-if="wallet.lastBlockStakeSpent">
            <td>Last Block Stake Spend</td>
            <td>
              @ Block:
              <span
                class="clickable"
                v-on:click="routeToBlockPage(wallet.lastBlockStakeSpent.height)"
                >{{ wallet.lastBlockStakeSpent.height }}</span
              >
              Txid:
              <span
                class="clickable"
                v-on:click="routeToHashPage(wallet.lastBlockStakeSpent.txid)"
                >{{ wallet.lastBlockStakeSpent.txid }}</span
              >
            </td>
          </tr>
        </tbody>
      </v-simple-table>
      <br />
      <v-simple-table  v-if="wallet.multisigAddressess">
        <tbody>
          <tr v-for="(multisigAddress, index) in wallet.multisigAddressess">
            <td>Multisig Address #{{ index + 1 }}</td>
            <td>{{ multisigAddress }}</td>
          </tr>
        </tbody>
      </v-simple-table>
      <br />
      <div v-for="(coinOutput, idx) in wallet.coinOutputs" v-bind:key="idx">
        <CoinOutput :output="coinOutput"/>
        <br/>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { mapState } from 'vuex'
import { UNIT } from '../../common/config'
import CoinOutput from '../Common/CoinOutput.vue'
import { toLocalDecimalNotation } from '../../common/helpers'

@Component({
  props: ['wallet'],
  name: 'UnlockHash',
  components: {
    CoinOutput
  },
  watch: {
    '$route.params.block' (val) {
      // call the method which loads your initial state
      this.$store.dispatch('SET_BLOCK_HEIGHT', val)
    },
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
      this.$store.dispatch('SET_BLOCK_HEIGHT', val)
      this.$router.push('/block/' + val)
    },
    renderValue: function (value: any) {
      return `${toLocalDecimalNotation(value)} ${UNIT}`
    }
  }
})
export default class UnlockHash extends Vue {
  unit = UNIT

  created () {
    window.scrollTo(0, 0)

    // If users navigates, recalculate lists
    this.$router.afterEach((newLocation: any) => {
      const hash = newLocation.params.hash
      this.$store.dispatch('SET_HASH', hash)
    })
  }
}
</script>
<style scoped>
.container h1 {
  text-align: left;
  font-size: 30px;
}
</style>
