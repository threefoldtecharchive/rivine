<template>
  <Fragment>
    <SingleSignatureFulfillment :fulfillment="fulfillment" v-if="fulfillment.getFulfillmentType() === FulfillmentType.SingleSignatureFulfillment"/>
    <AtomicSwapFulfillment :fulfillment="fulfillment" v-if="fulfillment.getFulfillmentType() === FulfillmentType.AtomicSwapFulfillment"/>
    <MultiSigFulfillment :fulfillment="fulfillment" v-if="fulfillment.getFulfillmentType() === FulfillmentType.MultisignatureFulfillment"/>
  </Fragment>
</template>
<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { FulfillmentType } from 'rivine-ts-types'
import SingleSignatureFulfillment from './SingleSignatureFulfillment.vue'
import AtomicSwapFulfillment from './AtomicSwapFulfillment.vue'
import MultiSigFulfillment from './MultiSigFulfillment.vue'
import { Fragment } from 'vue-fragment'

export default {
  data () {
    return {
      FulfillmentType
    }
  },
  props: ['fulfillment'],
  components: {
    SingleSignatureFulfillment,
    AtomicSwapFulfillment,
    MultiSigFulfillment,
    Fragment
  },
  name: 'Fulfillment'
}
</script>