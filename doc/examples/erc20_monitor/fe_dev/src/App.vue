<template>
  <div id="app">
    <TokenBalance name="TTFT20" contract-addr="0xd4466DAb724DD9Dc860e08E48a78D20e75361A25" />
  </div>
</template>

<script>
import TokenBalance from "./components/TokenBalance.vue";

export default {
  name: "app",
  components: {
    TokenBalance
  }
};
</script>

<style>
#app {
  font-family: "Avenir", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
