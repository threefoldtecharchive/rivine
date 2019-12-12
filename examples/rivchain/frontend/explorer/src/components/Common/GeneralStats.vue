<template>
  <div>
    <sui-table single-line>
      <sui-table-header>
        <sui-table-row>
          <sui-table-header-cell text-align="center"
            >General Statistics</sui-table-header-cell
          >
          <sui-table-header-cell></sui-table-header-cell>
        </sui-table-row>
      </sui-table-header>
      <sui-table-body>
        <sui-table-row>
          <sui-table-cell>Current Height</sui-table-cell>
          <sui-table-cell>{{ getExplorer().height }}</sui-table-cell>
        </sui-table-row>
        <sui-table-row>
          <sui-table-cell>Current Block</sui-table-cell>
          <sui-table-cell>{{ getExplorer().blockid }}</sui-table-cell>
        </sui-table-row>
        <sui-table-row>
          <sui-table-cell>Difficulty</sui-table-cell>
          <sui-table-cell>{{ getExplorer().difficulty }}</sui-table-cell>
        </sui-table-row>
      </sui-table-body>
    </sui-table>

    <sui-card>
      <sui-card-header text-align="center">General Statistics</sui-card-header>
      <sui-card-content>
        <div class="container">
          <ul class="list">
            <li>Height</li>
            <li>Block Id</li>
            <li>Difficulty</li>
          </ul>
          <ul class="list">
            <li v-for="item in getExplorer()" v-bind:key="item">
              <a>{{ item }}</a>
            </li>
          </ul>
        </div>
      </sui-card-content>
    </sui-card>
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import { pick } from 'lodash'

@Component
export default class GeneralStats extends Vue {
  getExplorer () {
    return pick(this.$store.getters.EXPLORER, [
      'height',
      'blockid',
      'difficulty'
    ])
  }

  created () {
    this.$store.dispatch('SET_EXPLORER')
  }
}
</script>

<style scoped>
.container {
  display: flex;
  justify-content: space-between;
}
.list {
  list-style: none;
}
.category {
  font-weight: bold;
  font-size: 1rem;
}
</style>
