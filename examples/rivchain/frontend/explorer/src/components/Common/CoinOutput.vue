<template>
  <div>
    <v-simple-table>
      <thead>
        <tr>
          <th colspan="3" class="ten wide">Coin Output</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>Block Height</td>
          <td
            class="clickable"
            v-on:click="routeToBlockPage(output.blockHeight)"
          >{{ output.blockHeight }}</td>
        </tr>
        <tr>
          <td>Transaction ID</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(output.txId)"
          >{{ output.txId }}</td>
        </tr>
        <tr>
          <td>ID</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(output.id)"
          >{{ output.id }}</td>
        </tr>
        <Condition :condition="output.condition" v-if="output.condition"/>
        <tr>
          <td>Value</td>
          <td>{{ renderValue(output.value) }}</td>
        </tr>

        <tr v-if="output.creationTime">
          <td>Creation Time</td>
          <td>{{ formatReadableDate(output.creationTime) }}</td>
        </tr>

        <tr v-if="output.creationValue">
          <td>Creation Value</td>
          <td>{{ output.creationValue }}</td>
        </tr>

        <tr v-if="output.feeComputationTime">
          <td>Current Age</td>
          <td>{{ formatTimeElapsed(output.feeComputationTime - output.creationTime) }}</td>
        </tr>

        <tr v-if="output.custodyFee">
          <td>Custody Fee Paid</td>
          <td>{{ renderValue(output.custodyFee) }}</td>
        </tr>

        <tr v-if="output.spendableValue">
          <td>Spendable Value</td>
          <td>{{ renderValue(output.spendableValue) }}</td>
        </tr>

        <tr>
          <td>Has been spent</td>
          <td>{{ output.spent ? 'Yes' : 'No' }}</td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>
<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { UnlockhashCondition, Currency } from 'rivine-ts-types'
import { PRECISION, UNIT } from '../../common/config'
import Condition from '../Conditions/Condition.vue'
import { toLocalDecimalNotation, formatTimeElapsed, formatReadableDate } from '../../common/helpers'

@Component({
  props: ['output'],
  name: 'CoinOutput',
  components: {
    Condition
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
      const v = toLocalDecimalNotation(value)
      return `${toLocalDecimalNotation(value)} ${UNIT}`
    },
    formatTimeElapsed,
    formatReadableDate
  }
})
export default class CoinOuput extends Vue {}
</script>