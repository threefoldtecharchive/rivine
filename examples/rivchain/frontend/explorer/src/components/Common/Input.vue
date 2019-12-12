<template>
  <div>
    <v-simple-table >
      <thead>
        <tr>
          <th colspan="3" class="ten wide">Used output</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>ID</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(input.parentid)"
          >{{ input.parentid }}</td>
        </tr>

        <Condition :condition="input.parentOutput.condition" />


        <tr v-if="input.creationTime">
          <td>Creation Time</td>
          <td>{{ formatReadableDate(input.creationTime) }}</td>
        </tr>

        <tr v-if="input.creationValue">
          <td>Creation Value</td>
          <td>{{ input.creationValue }}</td>
        </tr>

        <tr v-if="input.feeComputationTime">
          <td>Current Age</td>
          <td>{{ formatTimeElapsed(input.feeComputationTime - input.creationTime) }}</td>
        </tr>

        <tr v-if="input.custodyFee">
          <td>Custody Fee Paid</td>
          <td>{{ renderValue(input.custodyFee) }}</td>
        </tr>

        <tr v-if="input.spendableValue">
          <td>Spendable Value</td>
          <td>{{ renderValue(input.spendableValue) }}</td>
        </tr>

        <tr>
          <td>Value</td>
          <td>{{ renderValue(input.parentOutput.value) }}</td>
        </tr>
      </tbody>
    </v-simple-table>

    <br/>
    <Fulfillment :fulfillment="input.fulfillment" />

  </div>
</template>
<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { UnlockhashCondition, Currency } from 'rivine-ts-types'
import { PRECISION, UNIT } from '../../common/config'
import Fulfillment from '../Fulfillments/Fulfillment.vue'
import Condition from '../Conditions/Condition.vue'
import { toLocalDecimalNotation, formatReadableDate, formatTimeElapsed } from '../../common/helpers'

@Component({
  props: ['input'],
  name: 'Input',
  components: {
    Fulfillment,
    Condition
  },
  methods: {
    routeToHashPage: function (val) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    },
    renderValue: function (value: any) {
      return `${toLocalDecimalNotation(value)} ${UNIT}`
    },
    formatReadableDate,
    formatTimeElapsed
  }
})
export default class Input extends Vue {}
</script>