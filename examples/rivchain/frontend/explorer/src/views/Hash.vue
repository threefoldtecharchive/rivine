<template lang="html">
  <div>
    <navigation />

    <v-content>
      <v-container
        class="fill-height"
        fluid
      >

        <div class="searchBar">
          <Search category="hash" description="Hash" />
        </div>

        <v-skeleton-loader
          v-if="this.$store.getters.LOADING === true"
          type="table"
          min-width="75vw"
        ></v-skeleton-loader>

        <div v-else>
          <BlockstakeOutputHash class="container" v-if="this.$store.getters.HASH.kind() === responseType.BlockstakeOutputInfo" />

          <CoinOutputHash class="container" v-else-if="this.$store.getters.HASH.kind() === responseType.CoinOutputInfo"/>

          <Wallet class="container" :wallet="this.$store.getters.HASH" v-else-if="this.$store.getters.HASH.kind() === responseType.Wallet"/>

          <Transaction class="container" :transaction="this.$store.getters.HASH" v-else-if="this.$store.getters.HASH.kind() === responseType.Transaction"/>
        </div>

      </v-container>
    </v-content>

  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import BlockstakeOutputHash from '../components/Outputs/BlockstakeOutputHash.vue'
import CoinOutputHash from '../components/Outputs/CoinOutputHash.vue'
import Transaction from '../components/Transactions/Transaction.vue'
import Wallet from '../components/Wallets/Wallet.vue'
import Navigation from '../components/Common/Navigation.vue'
import Search from '../components/Common/Search.vue'
import Fragment from 'vue-fragment'
import { ResponseType, TransactionType } from 'rivine-ts-types'

@Component({
  name: 'Hash',
  components: {
    BlockstakeOutputHash,
    CoinOutputHash,
    Transaction,
    Wallet,
    Navigation,
    Search,
    Fragment
  },
  watch: {
    '$store.state.block': function () {
      this.$router.push('/block/' + this.$store.state.block.block.height)
    }
  }
})
export default class Hash extends Vue {
  loading: boolean = false
  responseType = ResponseType
  transactionType = TransactionType

  created () {
    if (!this.$route.params.hash) {
      this.$router.push('/')
    }
    if (!this.$store.getters.HASH.hashtype) {
      this.loading = true
      this.$store.dispatch('SET_HASH', this.$route.params.hash).then(() => {
        this.loading = false
      })
    }

    if (this.$store.getters.HASH === '') {
      this.$router.push('/notfound')
    }
  }
}
</script>
<style scoped>
.spinner {
  margin: "auto";
  margin-top: 50px;
  height: 500px;
}
.margin {
  margin-top: 20px;
  margin-bottom: 20px;
}
</style>
