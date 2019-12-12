<template lang="html">
  <div>
    <h1>Hash</h1>
    <v-simple-table >
      <thead>
        <tr>
          <th colspan="3">BlockStake Output</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>ID</td>
          <td>{{ this.$route.params.hash }}</td>
        </tr>

        <tr>
          <td>Transaction ID</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(blockStakeOutputInfo.output.txId)"
          >
            {{ blockStakeOutputInfo.output.txId }}
          </td>
        </tr>

        <tr>
          <td>Address</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(unlockhash)"
          >
            {{ unlockhash }}
          </td>
        </tr>

        <tr>
          <td>Value</td>
          <td>{{ blockStakeOutputInfo.output.value }}</td>
        </tr>

        <tr>
          <td>Has been spent</td>
          <td v-if="blockStakeOutputInfo.input">Yes</td>
          <td v-else>No</td>
        </tr>
      </tbody>
    </v-simple-table>
    <br/>
    <v-simple-table  v-if="blockStakeOutputInfo.input">
      <thead>
        <tr>
          <th colspan="3">BlockStake Input</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>ID</td>
          <td>{{ this.$route.params.hash }}</td>
        </tr>

        <tr>
          <td>Transaction ID</td>
          <td
            class="clickable"
            v-on:click="routeToHashPage(blockStakeOutputInfo.input.txId)"
          >
            {{ blockStakeOutputInfo.input.txId }}
          </td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { mapState } from 'vuex'
import { BlockstakeOutputInfo } from 'rivine-ts-types'
import { getUnlockHash } from '../../common/helpers'

@Component({
  name: 'BlockstakeOutputHash',
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
      this.$router.push('/block/' + val)
    }
  }
})
export default class BlockstakeOutputHash extends Vue {
  blockStakeOutputInfo?: BlockstakeOutputInfo
  isLoading: boolean = false
  unlockhash?: string

  created () {
    window.scrollTo(0, 0)
    this.blockStakeOutputInfo = this.$store.getters.HASH as BlockstakeOutputInfo
    this.unlockhash = getUnlockHash(this.blockStakeOutputInfo)
    this.isLoading = true
    // If users navigates, recalculate lists
    this.$router.afterEach((newLocation: any) => {
      const hash = newLocation.params.hash
      this.$store.dispatch('SET_HASH', hash).then(() => {
        this.blockStakeOutputInfo = this.$store.getters.HASH as BlockstakeOutputInfo
        this.unlockhash = getUnlockHash(this.blockStakeOutputInfo)
      })
    })
    this.isLoading = false
  }
}
</script>
<style scoped>
h1 {
  text-align: left;
  font-size: 30px;
}
</style>
