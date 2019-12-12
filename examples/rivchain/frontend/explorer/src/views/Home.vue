<template>
  <v-app id="inspire">
    <Navigation />

    <v-content>
      <v-container
        class="fill-height"
        fluid
      >
        <v-row
          align="center"
          justify="center"
        >
          <div class="header">
            <h1>{{ Name }} Explorer</h1>
            <h2>Search any hash or block.</h2>
            <div>
              <Search />
            </div>
          </div>
        </v-row>
      </v-container>
    </v-content>
  </v-app>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Search from '../components/Common/Search.vue'
import Navigation from '../components/Common/Navigation.vue'
import { NAME } from '../common/config'

@Component({
  components: {
    Search,
    Navigation
  }
})
export default class Home extends Vue {
  Name = NAME
  SearchVal: string = ''
}
</script>
<style scoped>
.header {
  width: 60%;
  margin: auto;
}
</style>
