<template>
  <v-simple-table >
    <thead>
      <tr>
        <th colspan="3">Multi Signature Fulfillment</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Type</td>
        <td>{{ fulfillment.type }}</td>
      </tr>
      <Fragment v-for="(keypair, index) in fulfillment.pairs" v-bind:key="index">
        <tr>
          <td>Publickey #{{ index + 1 }}</td>
          <td>{{ keypair.publickey }}</td>
        </tr>
        <tr>
          <td>Signature #{{ index + 1 }}</td>
          <td>{{ keypair.signature }}</td>
        </tr>
      </Fragment>
    </tbody>
  </v-simple-table>
</template>
<script lang="ts">
import { Fragment } from 'vue-fragment'
export default {
  props: ['fulfillment'],
  components: {
    Fragment
  }
}
</script>