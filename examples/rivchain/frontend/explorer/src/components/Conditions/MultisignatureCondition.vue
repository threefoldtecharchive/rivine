<template>
  <Fragment>
    <tr>
      <td>Multisig Address</td>
      <td
        class="clickable"
        v-on:click="routeToHashPage(condition.multisigAddress)"
      >{{ condition.multisigAddress }}</td>
    </tr>

    <Fragment v-for="(address, index) in condition.unlockhashes" v-bind:key="index">
      <tr>
        <td>Unlock Hash #{{ index + 1 }}</td>
        <td
          class="clickable"
          v-on:click="routeToHashPage(address)"
        >{{ address }}</td>
      </tr>
    </Fragment>

    <tr>
      <td>Minimum Signature Count</td>
      <td>{{ condition.signatureCount }}</td>
    </tr>
  </Fragment>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Fragment } from 'vue-fragment'

@Component({
  props: ['condition'],
  name: 'MultisigCondition',
  methods: {
    routeToHashPage: function (val) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    }
  },
  components: {
    Fragment
  }
})
// Export as class because Vue will understand this.$store etc..
export default class MultisigCondition extends Vue {}
</script>