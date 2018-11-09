<template>
  <div id="app">
    <TokenBalance name="ERC20-TTFT" contract-addr="0xe392e4dadeef43f326370d9e9b9f4990bf2636f5" />
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
