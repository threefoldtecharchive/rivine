<template>
  <v-form v-on:submit.prevent="handleSearch">
    <div class="ui icon input searchBar">
      <i class="search icon"></i>
      <v-text-field
        v-model="SearchVal"
        type="search"
        name="search"
        :error="this.error"
        :placeholder="getPlaceHolder()"
        v-on:keyup.enter="handleSearch"
        icon-pack="fas"
        icon="search"
        v-validate="'required'"
      />
    </div>
    <p v-if="error">{{ error }}</p>
  </v-form>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component, Prop } from 'vue-property-decorator'

@Component({})
export default class Search extends Vue {
  SearchVal: string = ''
  error: string = ''

  @Prop({ default: 'all' })
  category!: string
  @Prop({ default: 'Blocks, Transactions...' })
  description!: string

  getPlaceHolder () {
    return 'Search ' + this.description
  }

  handleSearch () {
    this.checkSearchInput()
  }

  checkSearchInput () {
    if (this.SearchVal.length > 60 || this.category === 'hash') {
      this.$store.dispatch('SET_HASH', this.SearchVal)
      this.$router.push('/hashes/' + this.SearchVal)
    } else if (parseInt(this.SearchVal) || this.category === 'block') {
      this.$store.dispatch('SET_BLOCK_HEIGHT', this.SearchVal)
      this.$router.push('/block/' + this.SearchVal)
    } else {
      this.$store.dispatch('SET_ERROR', `${this.SearchVal}`)
      this.$router.push('/notfound')
    }
  }

  created () {
    this.$store.dispatch('SET_EXPLORER', this.SearchVal)
  }
}
</script>

<style scoped>
.searchBar {
  width: 100%;
}
.searchButton {
  margin-top: 5%;
}
</style>
