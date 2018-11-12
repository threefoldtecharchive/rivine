<template>
  <div class="token-balance">
    <h3 class="title">{{ name }}</h3>
    <div class="container">
        <div class="address-input-container">
            <span class="address-title">Address:</span>
            <input class="address-input" v-model="address" @keypress.enter="getBalance">
        </div>
        <span class="balance">{{ balance | decimal }} {{ name }}</span>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
export default {
  name: "TokenBalance",
  props: {
    name: String,
    contractAddr: String
  },
  data: function() {
      return {
        balance: 0,
        address: '',
        interval: null
      }
  },
  filters: {
      decimal(value) {
          // most erc20 contracts have 18 decimals;
          const decimals = 1000000000000000000;
          return (value / decimals).toFixed(5)
      }
  },
  methods: {
      getBalance: function() {
          if (this.interval) {
              clearInterval(this.interval);
              this.interval = null;
          }
          var url = '/tokenbalance?address=' + this.address + '&contractaddress=' + this.contractAddr;
          axios.get(url).then(response => (this.balance = response.data));
          // repeat check every second
          this.interval = setInterval(() => {axios.get(url).then(response => this.balance = response.data)}, 1000);
      }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
    margin: 20px 0px;
}
.title {
    align-self: flex-start;
    padding: 0px 25px;
}
.token-balance {
    display: flex;
    flex: 1;
    width: 80%;
    justify-content: initial;
    flex-direction: column;
    padding-left: 25px;
    margin: 25px 0px;
}
.container {
    display: flex;
    flex: 1;
    flex-flow: row;
    justify-content: space-between;
    border: 2px solid darkcyan;
    border-radius: 25px;
    padding-right: 25px;
    padding-left: 15px;
    background-image: linear-gradient(to bottom right, lightblue, lightgreen);
}
.address-input-container {
    align-self: center;
    display: flex;
}
.address-input-container > * {
    margin: 10px;
}
.address-input {
    background-color: transparent;
    border: 1px solid cornflowerblue;
    border-radius: 10px;
    padding-left: 5px;
    padding-right: 5px;
    width: 300px;
}
.balance {
    align-self: center;
}
</style>