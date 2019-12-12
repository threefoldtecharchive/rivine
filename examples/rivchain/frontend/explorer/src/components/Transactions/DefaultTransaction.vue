<template>
  <div>
    <br/>
    <div>
      <div v-if="transaction.blockStakeInputs" >
        <h3 v-if="transaction.blockStakeInputs.length > 0">Blockstake Inputs</h3>
        <div
          v-for="(input, index) in transaction.blockStakeInputs"
          v-bind:key="index"
        >
          <Input :input="input" />
          <br/>
        </div>
      </div>

      <div v-if="transaction.blockStakeOutputs" >
        <h3 v-if="transaction.blockStakeOutputs.length > 0">Blockstake Outputs</h3>

        <div
          v-for="(output, index) in transaction.blockStakeOutputs"
          v-bind:key="index"
        >
          <Output :output="output" />
          <br/>
        </div>
      </div>

      <div v-if="transaction.coinInputs" >
        <h3 v-if="transaction.coinInputs.length > 0">Coin Inputs</h3>
        <div
          v-for="(input, index) in transaction.coinInputs"
          v-bind:key="index"
        >
          <Input :input="input" />
          <br/>
        </div>
      </div>

      <div v-if="transaction.coinOutputs" >
        <h3 v-if="transaction.coinOutputs.length > 0">Coin Outputs</h3>
        <div
          v-for="(output, index) in transaction.coinOutputs"
          v-bind:key="index"
        >
          <Output :output="output" />
          <br/>
        </div>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Input from '../Common/Input.vue'
import Output from '../Common/Output.vue'

@Component({
  props: ['transaction'],
  name: 'DefaultTransaction',
  components: {
    Input,
    Output
  }
})
export default class DefaultTransaction extends Vue {}
</script>