<template>
  <div>
    <v-simple-table >
      <thead>
        <tr>
          <th colspan="3">Transaction version {{ transaction.version }}</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>Block Height</td>
          <td v-if="transaction.blockHeight == 0">
            unconfirmed
          </td>
          <td
            v-else
            class="clickable"
            v-on:click="routeToBlockPage(transaction.blockHeight)"
          >
            {{ toLocalDecimalNotation(transaction.blockHeight) }}
          </td>
        </tr>
        <tr v-if="transaction.blockHeight !== 0">
          <td>Confirmations</td>
          <td>
            {{ toLocalDecimalNotation(this.$store.getters.EXPLORER.height - transaction.blockHeight + 1) }}
          </td>
        </tr>
        <tr>
          <td>ID</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(transaction.id)"
          >{{ transaction.id }}</td>
        </tr>
        <tr v-if="transaction.coinInputs.length > 0">
          <td>Coin Input Count</td>
          <td>
            {{ transaction.coinInputs.length }}
          </td>
        </tr>
        <tr v-if="transaction.coinOutputs.length > 0">
          <td>Coin Output Count</td>
          <td>
            {{ transaction.coinOutputs.length }}
          </td>
        </tr>
        <tr v-if="transaction.blockStakeInputs.length > 0">
          <td>Blockstake Input Count</td>
          <td>
            {{ transaction.blockStakeInputs.length }}
          </td>
        </tr>
        <tr v-if="transaction.blockStakeOutputs.length > 0">
          <td>Blockstake Output Count</td>
          <td>
            {{ transaction.blockStakeOutputs.length }}
          </td>
        </tr>
        <tr v-if="transaction.arbitrarydata">
          <td>Arbitrary Data Byte</td>
          <td>
            {{ transaction.arbitrarydata.length }}
          </td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>
<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { toLocalDecimalNotation } from '../../common/helpers'

export default {
  props: ['transaction'],
  methods: {
    routeToHashPage: function (val: string) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    },
    routeToBlockPage: function (val) {
      this.$store.dispatch('SET_BLOCK_HEIGHT', val)
      this.$router.push('/block/' + val)
    },
    toLocalDecimalNotation
  },
  created () {
    window.scrollTo(0, 0)
  },
  name: 'TransactionSummary'
}
</script>
<style scoped>
</style>