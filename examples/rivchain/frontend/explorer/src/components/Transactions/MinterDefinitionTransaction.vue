<template>
  <div>
    <div>
      <h2>Minter Definition Fulfillment</h2>
      <Fulfillment :fulfillment="transaction.minterDefinitionFulfillment" />
    </div>
    <div>
      <h2>New Mint Condition</h2>
      <v-simple-table >
        <Condition :condition="transaction.minterDefinitionCondition" />
      </v-simple-table>
    </div>
  </div>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Fulfillment from '../Fulfillments/Fulfillment.vue'
import Condition from '../Conditions/Condition.vue'

@Component({
  props: ['transaction'],
  name: 'MinterDefinitionTransaction',
  components: {
    Fulfillment,
    Condition
  }
})
export default class MinterDefinitionTransaction extends Vue {}
</script>