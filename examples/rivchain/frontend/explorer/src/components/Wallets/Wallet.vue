<template>
  <div class="container">
    <BlockCreatorWallet :wallet="wallet" v-if="wallet.isBlockCreator"/>
    <CustodyVoidWallet :wallet="wallet" v-else-if="wallet.isCustodyVoid"/>
    <DefaultWallet :wallet="wallet" v-else/>
  </div>
</template>
<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { TransactionType } from 'rivine-ts-types'
import { PRECISION, UNIT } from '../../common/config'
import BlockCreatorWallet from './BlockCreatorWallet.vue'
import DefaultWallet from './DefaultWallet.vue'
import CustodyVoidWallet from './CustodyVoidWallet.vue'

export default {
  data () {
    return {
      transactionType: TransactionType
    }
  },
  props: ['wallet'],
  components: {
    BlockCreatorWallet,
    DefaultWallet,
    CustodyVoidWallet
  },
  name: 'wallet'
}
</script>