<template>
  <Fragment>
    <tr>
      <td>Contract Address</td>
      <td>
        {{ this.$route.params.hash }}
      </td>
    </tr>

    <tr>
      <td>Sender</td>
      <td
        class="clickable"
        v-on:click="routeToHashPage(condition.sender)"
      >{{ condition.sender }}</td>
    </tr>

    <tr>
      <td>Receiver</td>
      <td
        class="clickable"
        v-on:click="routeToHashPage(condition.receiver)"
      >{{ condition.receiver }}</td>
    </tr>

    <tr>
      <td>Hashed Secret</td>
      <td>{{ condition.hashedSecret }}</td>
    </tr>

    <tr>
      <td>Timelock</td>
      <td>{{ condition.timelock }}</td>
    </tr>

    <tr>
      <td>Unlocked for refunding since</td>
      <td>{{ formatReadableDate(condition.timelock) }}</td>
    </tr>
  </Fragment>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { formatReadableDate } from '../../common/helpers'
import { Fragment } from 'vue-fragment'

@Component({
  props: ['condition'],
  name: 'AtomicSwapOutputTable',
  components: {
    Fragment
  },
  methods: {
    routeToHashPage: function (val) {
      this.$store.dispatch('SET_HASH', val)
      this.$router.push('/hashes/' + val)
    },
    formatReadableDate
  }
})
// Export as class because Vue will understand this.$store etc..
export default class AtomicSwapOutputTable extends Vue {}
</script>