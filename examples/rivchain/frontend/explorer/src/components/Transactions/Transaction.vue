<template>
  <div>
    <h1>Transaction</h1>
    <TransactionSummary :transaction="transaction"/>
    <DefaultTransaction
      :transaction="transaction"
      v-if="transaction.getTransactionType() === transactionType.DefaultTransaction"
    />
    <MinterDefinitionTransaction
      :transaction="transaction"
      v-if="transaction.getTransactionType() === transactionType.MinterDefinitionTransaction"
    />
    <CoinCreationTransaction
      :transaction="transaction"
      v-if="transaction.getTransactionType() === transactionType.CoinCreationTransaction"
    />
  </div>
</template>
<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { TransactionType } from 'rivine-ts-types'
import { PRECISION, UNIT } from '../../common/config'
import TransactionSummary from './TransactionSummary.vue'
import DefaultTransaction from './DefaultTransaction.vue'
import MinterDefinitionTransaction from './MinterDefinitionTransaction.vue'
import CoinCreationTransaction from './CoinCreationTransaction.vue'
import { toLocalDecimalNotation } from '../../common/helpers'

export default {
  data () {
    return {
      transactionType: TransactionType
    }
  },
  props: ['transaction'],
  components: {
    DefaultTransaction,
    MinterDefinitionTransaction,
    CoinCreationTransaction,
    TransactionSummary
  },
  methods: {
    routeToHashPage: function (val: string) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    },
    routeToBlockPage: function (val) {
      this.$store.dispatch('', val)
      this.$router.push('/block/' + val)
    },
    toLocalDecimalNotation
  },
  created () {
    window.scrollTo(0, 0)
  },
  name: 'Transaction'
}
</script>
<style scoped>
</style>