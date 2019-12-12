<template>
  <div>
    <navigation />

    <v-content>
      <v-container
        class="fill-height"
        fluid
      >
        <div class="searchBar">
          <Search category="block" description="Block Heights" />
        </div>
        <Block :block="this.$store.getters.BLOCK" />
      </v-container>
    </v-content>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Navigation from '../components/Common/Navigation.vue'
import Search from '../components/Common/Search.vue'
import Block from '../components/Blocks/Block.vue'
import Fragment from 'vue-fragment'

@Component({
  components: {
    Navigation,
    Search,
    Block,
    Fragment
  }
})
export default class Blocks extends Vue {
  created () {
    window.scrollTo(0, 0)
    this.$store.dispatch('SET_EXPLORER')
  }

  beforeCreate () {
    const { height } = this.$route.params
    if (height) {
      this.$store.dispatch('SET_BLOCK_HEIGHT', height)
    }
  }
}
</script>
