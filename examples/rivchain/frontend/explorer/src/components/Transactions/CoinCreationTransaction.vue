<template>
  <div>
    <div>
      <h2>Coin Creation Fulfillment</h2>
      <Fulfillment :fulfillment="transaction.coinCreationFulfillment" />
    </div>
    <div>
      <h2>Coin Creation Outputs</h2>
      <Fragment v-for="(output, index) in transaction.coinCreationOutputs" v-bind:key="index">
        <Output :output="output" />
        <br/>
      </Fragment>
    </div>
  </div>
</template>
<script lang="ts">
import { Fragment } from 'vue-fragment'
import { Component, Vue } from 'vue-property-decorator'
import Fulfillment from '../Fulfillments/Fulfillment.vue'
import Condition from '../Conditions/Condition.vue'
import Output from '../Common/Output.vue'

@Component({
  props: ['transaction'],
  name: 'CoinCreationTransaction',
  components: {
    Fulfillment,
    Condition,
    Output,
    Fragment
  }
})
export default class CoinCreationTransaction extends Vue {}
</script>