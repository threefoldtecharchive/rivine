<template lang="html">
  <div>
    <Navigation />
      <br />
      <v-content>
        <v-container
          class="fill-height"
          fluid
        >

        <v-row class="mb-6">
          <v-col
            v-for="n in 2"
            :key="n"
            :lg="6"
            :md="12"
            :sm="12"
            :cols="12"
          >
            <ChartInfo v-if="n === 1" :data="infoData" class="rowcontainer"/>

            <v-form v-if="n === 2" @submit.prevent="fetchData()" class="rowcontainer">
              <div class="field">
                <label>Show graphs of the latest blocks:</label>
                <v-text-field
                  v-model="history"
                  type="number"
                  name="blocks"
                  placeholder="block"
                />
              </div>
              <v-btn small color="primary" @click="fetchData">Search</v-btn>

              <div class="input-field">
                <div class="field">
                  <label>Show graphs of the specfied block range:</label>
                  <v-text-field
                    type="number"
                    v-model="historyFrom"
                    name="from"
                    placeholder="from block"
                  />
                  <v-text-field
                    type="number"
                    v-model="historyUntil"
                    name="until"
                    placeholder="until block"
                  />
                </div>
                <v-btn small color="primary" @click="fetchData">Search</v-btn>
              </div>
            </v-form>
          </v-col>
        </v-row>

        <v-row >
          <v-card
            class="mx-auto rowcontainer-chart"
          >
            <v-card-title>
              Chain Height
            </v-card-title>
            <LineChart
              v-if="!loading"
              :data="chainHeightData"
              :options="chainHeightOptions"
              :color="2"
            />
          </v-card>
        </v-row>

        <v-row>
          <v-card
            class="mx-auto rowcontainer-chart"
          >
            <v-card-title>
              Block Creation Time (seconds since previous block)
            </v-card-title>
            <LineChart
              v-if="!loading"
              :data="blockTimeData"
              :options="blockTimeOptions"
              :color="1"
            />
          </v-card>
        </v-row>

        <v-row>
          <v-card
            class="mx-auto rowcontainer-chart"
          >
            <v-card-title>
              Estimate Active Blockstakes
            </v-card-title>
            <LineChart
              v-if="!loading"
              :data="activeBsData"
              :options="activeBsOptions"
              :color="2"
            />
          </v-card>
        </v-row>

        <v-row>
          <v-card
            class="mx-auto rowcontainer-chart"
          >
            <v-card-title>
              Block Transaction Count
            </v-card-title>
            <LineChart
              v-if="!loading"
              :data="blockTransactionCountData"
              :options="blockTransactionCountOptions"
              :color="1"
            />
          </v-card>
        </v-row>

        <v-row>
          <v-card
            class="mx-auto rowcontainer-chart"
          >
            <v-card-title>
              Block Difficulty
            </v-card-title>
            <LineChart
              v-if="!loading"
              :data="blockDifficultyData"
              :options="blockDifficultyOptions"
              :color="2"
            />
          </v-card>
        </v-row>

        <v-row>
          <v-card
            class="mx-auto rowcontainer-chart"
          >
            <v-card-title>
             Block Creator Distribution
            </v-card-title>
            <PieChart
              v-if="!loading"
              :data="blockCreatorPieGraphData"
              :options="blockCreatorPieGraphOptions"
            />
          </v-card>
        </v-row>
      </v-container>
    </v-content>
  </div>
</template>

<script lang='ts'>
/* eslint:disable */
import axios from 'axios'
import { API_URL } from '../common/config'
import Navigation from '../components/Common/Navigation.vue'
import LineChart from '../components/Charts/LineChart.vue'
import PieChart from '../components/Charts/PieChart.vue'
import ChartInfo from '../components/Charts/ChartInfo.vue'
import { formatReadableDateForCharts } from '../common/helpers'
import { cloneDeep } from 'lodash'

const tooltipFormat = 'DD/MM/YYYY HH:MM'

export default {
  name: 'charts',
  data () {
    return {
      chainHeightData: {},
      chainHeightOptions: {},
      blockTimeData: {},
      blockTimeOptions: {},
      activeBsData: {},
      activeBsOptions: {},
      blockTransactionCountData: {},
      blockTransactionCountOptions: {},
      blockDifficultyData: {},
      blockDifficultyOptions: {},
      blockCreatorPieGraphData: {},
      blockCreatorPieGraphOptions: {},
      infoData: {},
      loadData: {},
      loading: false,
      gradient: {},
      gradient2: {},
      history: 75,
      historyFrom: '',
      historyUntil: '',
      defaultOptions: {
        scales: {
          yAxes: [
            {
              gridLines: {
                display: true
              }
            }
          ],
          xAxes: [
            {
              gridLines: {
                display: false
              }
            }
          ]
        },
        legend: {
          display: true
        },
        responsive: true,
        maintainAspectRatio: false
      },
      defaultPieOptions: {
        legend: {
          display: true
        },
        responsive: true,
        maintainAspectRatio: false
      }
    }
  },
  components: {
    Navigation,
    LineChart,
    PieChart,
    ChartInfo
  },
  mounted () {
    this.fetchData()
  },
  methods: {
    fetchData () {
      let url = API_URL + '/explorer/stats/history?history=' + this.history
      if (this.historyFrom && this.historyUntil) {
        url =
          API_URL +
          `/explorer/stats/range?start=${this.historyFrom}&end=${this.historyUntil}`
      }
      this.loading = true
      axios({ method: 'GET', url }).then(result => {
        this.loadData = result.data
        this.mapDataForChainHeightGraph(result.data)
        this.mapDataForBlockTimeGraph(result.data)
        this.mapDataForActiveBlockstakesGraph(result.data)
        this.mapDataForBlockTransactionsCountGraph(result.data)
        this.mapDataForDifficultyGraph(result.data)
        this.mapDataForBlockCreatorPieGraph(result.data)
        this.mapInfoData(result.data)
        this.loading = false
        this.historyFrom = ''
        this.historyUntil = ''
      })
    },
    mapDataForChainHeightGraph (result) {
      const labels = result.blocktimestamps.map(x =>
        formatReadableDateForCharts(x)
      )
      this.chainHeightData = {
        labels,
        datasets: [
          {
            data: result.blockheights,
            label: 'Chain Height',
            labelColor: 'white',
            borderColor: '#FC2525',
            pointBackgroundColor: 'red',
            borderWidth: 1,
            pointBorderColor: 'white',
            lineWidth: 0.1,
            pointRadius: 3
          }
        ]
      }
      this.chainHeightOptions = cloneDeep(this.defaultOptions)
      this.chainHeightOptions.scales.xAxes = [
        {
          type: 'time',
          time: {
            parser: tooltipFormat
          }
        }
      ]
      const _this = this
      this.chainHeightOptions.onClick = function (e, chartElement) {
        if (chartElement.length === 0) return
        const index = chartElement[0]._index
        const height = result.blockheights[index]
        _this.$store.dispatch('SET_BLOCK_HEIGHT', height)
        _this.$router.push('/block/' + height)
      }
    },
    mapDataForBlockTimeGraph (result) {
      const labels = result.blockheights
      this.blockTimeData = {
        labels,
        datasets: [
          {
            data: result.blocktimes,
            label: 'Block Creation Time',
            labelColor: 'white',
            borderColor: '#FC2525',
            pointBackgroundColor: 'black',
            borderWidth: 1,
            pointBorderColor: 'white',
            lineWidth: 0.1,
            pointRadius: 2
          }
        ]
      }
      this.blockTimeOptions = cloneDeep(this.defaultOptions)
      this.blockTimeOptions.scales.yAxes = [
        {
          gridLines: {
            display: true
          },
          ticks: {
            suggestedMin: -200
          }
        }
      ]
      const _this = this
      this.blockTimeOptions.onClick = function (e, chartElement) {
        if (chartElement.length === 0) return
        const index = chartElement[0]._index
        const height = result.blockheights[index]
        _this.$store.dispatch('SET_BLOCK_HEIGHT', height)
        _this.$router.push('/block/' + height)
      }
    },
    mapDataForActiveBlockstakesGraph (result) {
      const labels = result.blocktimestamps.map(x =>
        formatReadableDateForCharts(x)
      )
      this.activeBsData = {
        labels,
        datasets: [
          {
            data: result.estimatedactivebs,
            label: 'Active BS',
            labelColor: 'white',
            borderColor: '#FC2525',
            pointBackgroundColor: 'red',
            borderWidth: 1,
            pointBorderColor: 'white',
            lineWidth: 0.1,
            pointRadius: 2
          }
        ]
      }
      this.activeBsOptions = cloneDeep(this.defaultOptions)
      this.activeBsOptions.scales.xAxes = [
        {
          type: 'time',
          time: {
            parser: tooltipFormat
          }
        }
      ]
      const _this = this
      this.activeBsOptions.onClick = function (e, chartElement) {
        if (chartElement.length === 0) return
        const index = chartElement[0]._index
        const height = result.blockheights[index]
        _this.$store.dispatch('SET_BLOCK_HEIGHT', height)
        _this.$router.push('/block/' + height)
      }
    },
    mapDataForBlockTransactionsCountGraph (result) {
      const labels = result.blockheights
      this.blockTransactionCountData = {
        labels,
        datasets: [
          {
            data: result.blocktransactioncounts,
            label: 'Transaction Count',
            labelColor: 'white',
            borderColor: '#FC2525',
            pointBackgroundColor: 'black',
            borderWidth: 1,
            pointBorderColor: 'white',
            lineWidth: 0.1,
            spanGaps: true,
            pointRadius: 2
          }
        ]
      }
      this.blockTransactionCountOptions = cloneDeep(this.defaultOptions)
      this.blockTransactionCountOptions.scales.yAxes = [
        {
          gridLines: {
            display: true
          },
          ticks: {
            suggestedMin: -1,
            suggestedMax: 1,
            stepSize: 0.5
          }
        }
      ]
      const _this = this
      this.blockTransactionCountOptions.onClick = function (e, chartElement) {
        if (chartElement.length === 0) return
        const index = chartElement[0]._index
        const height = result.blockheights[index]
        _this.$store.dispatch('SET_BLOCK_HEIGHT', height)
        _this.$router.push('/block/' + height)
      }
    },
    mapDataForDifficultyGraph (result) {
      const labels = result.blockheights
      this.blockDifficultyData = {
        labels,
        datasets: [
          {
            data: result.difficulties,
            label: 'Difficulty',
            labelColor: 'white',
            borderColor: '#FC2525',
            pointBackgroundColor: 'red',
            borderWidth: 1,
            pointBorderColor: 'white',
            lineWidth: 0.1,
            pointRadius: 2
          }
        ]
      }
      this.blockDifficultyOptions = cloneDeep(this.defaultOptions)
      this.blockDifficultyOptions.scales.yAxes = [
        {
          gridLines: {
            display: true
          },
          ticks: {
            stepSize: 100000
          }
        }
      ]
      const _this = this
      this.blockDifficultyOptions.onClick = function (e, chartElement) {
        if (chartElement.length === 0) return
        const index = chartElement[0]._index
        const height = result.blockheights[index]
        _this.$store.dispatch('SET_BLOCK_HEIGHT', height)
        _this.$router.push('/block/' + height)
      }
    },
    mapDataForBlockCreatorPieGraph (result) {
      const creators = Object.values(result.creators)
      const addresses = Object.keys(result.creators)
      this.blockCreatorPieGraphData = {
        labels: addresses,
        datasets: [
          {
            data: creators,
            backgroundColor: ['#3e95cd', '#8e5ea2']
          }
        ]
      }
      this.blockCreatorPieGraphOptions = this.defaultPieOptions
    },
    mapInfoData (result) {
      const length = result.difficulties.length

      const difficulty = result.difficulties[length - 1]
      const currentBlock = result.blockheights[length - 1]
      const estimatedActiveBs = result.estimatedactivebs[length - 1]

      this.infoData = {
        difficulty,
        currentBlock,
        estimatedActiveBs
      }
    }
  }
}
</script>
<style scoped>
.input-field {
  margin-top: 20px;
}
.pie-chart {
  margin-left: auto;
  margin-right: auto;
  border-radius: 1%;
  padding: 30px;
  margin-top: 50px;
  margin-bottom: 50px;
  width: 50%;
  box-shadow: 0 2px 16px 0 rgba(0, 0, 0, 0.3);
}
.Chart__title {
  color: white;
  margin-bottom: rem(40);
  font-weight: 600;
  font-size: rem(16);
}
.Chart__title > span {
  font-weight: 400;
  color: color(robin-egg-blue);
  font-size: rem(16);
  margin-left: rem(25);
}
</style>
