<template>
  <div class="pusher">
    <div class="ui inverted vertical masthead center aligned segment">
      <navigation />
        <v-content>
          <v-container
            class="fill-height"
            fluid
          >
            <div class="ui text container notfound">
              <div v-for="err in getError()" v-bind:key="err">
                <h4>{{ err }}</h4>
                <br>
              </div>
              <h4>Were you looking for an identifier instead? Please take into account the following:
                  All transaction‐, Block‐, Coin Output‐ and Blockstake Ouput‐identifiers have a fixed length of 64 characters.
              </h4>
              <div class="searchBar">
                <search />
              </div>
            </div>
          </v-container>
        </v-content>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Navigation from '../components/Common/Navigation.vue'
import Search from '../components/Common/Search.vue'

@Component({
  components: {
    Navigation,
    Search
  },
  methods: {
    getError () {
      const err = this.$store.getters.ERROR
      if (err instanceof String) {
        if (err.startsWith('01')) {
          return [
            `Invalid hash: ${err}`,
            `As the address starts with '01', you might be looking for a Wallet,
            please make sure the address is correct.`,
            `The wallet might not have received any coin or block stake outputs yet or be referenced by a multi signature wallet, in which case it is not visible in this explorer. Once a coin or block stake has been received into this wallet or a multi signature wallet references it, it will show up.`
          ]
        } else if (err.startsWith('03')) {
          return [
            `Invalid hash: ${err}`,
            `As the address starts with '03', you might be looking for a Multisig Wallet,
            please make sure the address is correct.`,
            `The MultiSig Wallet might not have received any coin or block stake outputs yet, in which case it is not visible in this explorer. Once a coin or block stake has been received into this wallet, it will show up..`
          ]
        } else {
          return [`Hash or Block with height: ${err} not found`]
        }
      }
    }
  }
})
export default class Error extends Vue {
  mounted () {
    const _this = this
    window.onpopstate = function () {
      _this.$router.push('/')
    }
  }

  beforeDestroy () {
    const _this = this
    // tslint:disable-next-line
    window.onpopstate = function () {}
  }
}
</script>
<style scoped>
.notfound {
  margin-top: 200px;
  text-align: center;
}
.searchBar {
  margin-top: 50px;
  margin-left: auto;
  margin-right: auto;
}
</style>
