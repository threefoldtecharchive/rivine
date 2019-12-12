<template lang="html">
  <div>
    <h1>Hash</h1>
    <div>
      <v-simple-table >
        <thead>
          <tr>
            <th colspan="3">Custody Fee Void</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Void Address</td>
            <td>{{ wallet.address }}</td>
          </tr>
          <tr>
            <td>Confirmed Custody Fees Paid</td>
            <td>{{ renderValue(wallet.confirmedCoinBalance) }}</td>
          </tr>
        </tbody>
      </v-simple-table>

      <br/>
      <div v-for="(coinOutput, idx) in wallet.coinOutputs" v-bind:key="idx">
        <CoinOutput :output="coinOutput"/>
        <br/>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { UNIT } from '../../common/config'
import CoinOutput from '../Common/CoinOutput.vue'
import { toLocalDecimalNotation } from '../../common/helpers'

@Component({
  props: ['wallet'],
  components: {
    CoinOutput
  },
  name: 'CustodyVoidWallet',
  methods: {
    renderValue: function (value: any) {
      return `${toLocalDecimalNotation(value)} ${UNIT}`
    }
  },
  created () {
    window.scrollTo(0, 0)
  }
})
export default class CustodyVoidWallet extends Vue {}
</script>
