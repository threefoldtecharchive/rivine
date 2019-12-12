<template>
  <tr>
    <td>Custody Fee Void Address</td>
    <td
      class="clickable"
      v-on:click="routeToHashPage(condition.custodyFeeVoidAddress)"
    >{{ condition.custodyFeeVoidAddress }}</td>
  </tr>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component({
  props: ['condition'],
  name: 'CustodyFeeCondition',
  methods: {
    routeToHashPage: function (val) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    }
  }
})
// Export as class because Vue will understand this.$store etc..
export default class CustodyFeeCondition extends Vue {}
</script>