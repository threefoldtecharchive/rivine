<template>
  <v-simple-table >
    <thead>
      <tr>
        <th colspan="3">Atomic Swap Fulfillment</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Type</td>
        <td>{{ fulfillment.type }}</td>
      </tr>
      <tr>
        <td>Publickey</td>
        <td>{{ fulfillment.publickey }}</td>
      </tr>
      <tr>
        <td>Signature</td>
        <td>{{ fulfillment.signature }}</td>
      </tr>
      <tr v-if="fulfillment.secret">
        <td>Hashed Secret</td>
        <td>{{ fulfillment.secret }}</td>
      </tr>
    </tbody>
  </v-simple-table>
</template>
<script lang="ts">
export default {
  props: ['fulfillment']
}
</script>