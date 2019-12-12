<template>
  <div>
    <v-simple-table>
      <thead>
        <tr>
          <th colspan="3" class="text-left">Blockstake Output</th>
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
        <Condition :condition="output.condition" />
        <tr>
          <td>Value</td>
          <td>{{ renderValue(output.value) }}</td>
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
import { PRECISION } from '../../common/config'
import Condition from '../Conditions/Condition.vue'
import { toLocalDecimalNotation } from '../../common/helpers'

@Component({
  props: ['output'],
  name: 'BlockStakeOutput',
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
      return `${toLocalDecimalNotation(value)} BS`
    }
  }
})
export default class BlockStakeOutput extends Vue {}
</script>