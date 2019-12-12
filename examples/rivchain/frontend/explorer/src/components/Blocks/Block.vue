<template lang="html">
  <div class="container">
    <h1>Block</h1>
    <v-simple-table class="table">
      <thead>
        <tr>
          <th colspan="3" class="text-left">Block statistics</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>Block Height</td>
          <td>
            {{ toLocalDecimalNotation(block.height) }}
          </td>
        </tr>
        <tr>
          <td>ID</td>
          <td
            class="clickable"
            v-on:click="routeToBlockPage(block.id)"
          >
            {{ block.id }}
          </td>
        </tr>
        <tr>
          <td>Confirmations</td>
          <td>
            {{ toLocalDecimalNotation(this.$store.getters.EXPLORER.height) }}
          </td>
        </tr>
        <tr>
          <td>Previous Block</td>
          <td
            class="clickable"
            v-on:click="routeToBlockPage(block.parentId)"
          >
            {{ block.parentId }}
          </td>
        </tr>
        <tr>
          <td>Time</td>
          <td>{{ formatBlockDate(block.timestamp) }}</td>
        </tr>
        <tr>
          <td>Active Blockstake</td>
          <td>
            {{ toLocalDecimalNotation(block.estimatedActiveBlockStakes) }} BS
          </td>
        </tr>
      </tbody>
    </v-simple-table>

    <div v-if="block.minerPayouts">
      <h2> Block creator rewards </h2>
      <div
        v-for="(output, index) in block.minerPayouts"
        v-bind:key="index"
      >
        <MinerPayout :output="output" class="table"/>
      </div>
    </div>

    <div v-if="block.transactions">
      <h2> Transactions </h2>
      <div
        v-for="(tx, index) in block.transactions"
        v-bind:key="index"
      >
        <TransactionSummary :transaction="tx" class="table"/>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { mapState } from 'vuex'
import { toLocalDecimalNotation } from '../../common/helpers'
import TransactionSummary from '../Transactions/TransactionSummary.vue'
import MinerPayout from '../Common/MinerOutput.vue'

@Component({
  name: 'Block',
  props: ['block'],
  components: {
    TransactionSummary,
    MinerPayout
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
      this.$store.dispatch('SET_HASH', val)
    }
  }
})
export default class Block extends Vue {
  blockDate = ''
  toLocalDecimalNotation = toLocalDecimalNotation

  created () {
    window.scrollTo(0, 0)
    if (
      !this.$route.params.block ||
      isNaN(parseInt(this.$route.params.block, 10))
    ) {
      this.$router.push('/blocks/')
    }
    if (!this.$store.getters.BLOCK.block) {
      this.$store.dispatch('SET_BLOCK_HEIGHT', this.$route.params.block)
    }
  }

  formatBlockDate (timestamp) {
    const blockDate = new Date(timestamp * 1000)
    const day = blockDate.getDate()
    const month = blockDate.toLocaleString('default', { month: 'long' })
    const year = blockDate.getFullYear()
    const hours = blockDate.getHours()
    const tempMinutes = blockDate.getMinutes()
    const minutes = tempMinutes < 10 ? `0${tempMinutes}` : tempMinutes

    return `${hours}:${minutes}, ${month} ${day}, ${year}`
  }
}
</script>
<style scoped>
.table {
  text-align: left;
  margin-top: 20px;
  margin-bottom: 20px;
}
.container h2 {
  text-align: left;
  font-size: 26px;
}
</style>
