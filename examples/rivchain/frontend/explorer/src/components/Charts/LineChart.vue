<script>
import { Line } from 'vue-chartjs'
import { first } from 'lodash'

export default {
  extends: Line,
  props: ['data', 'options', 'color'],
  mounted() {
    this.gradient = this.$refs.canvas
      .getContext('2d')
      .createLinearGradient(0, 0, 0, 450)
    this.gradient2 = this.$refs.canvas
      .getContext('2d')
      .createLinearGradient(0, 0, 0, 450)
    this.gradient.addColorStop(0, 'rgba(255, 0,0, 0.5)')
    this.gradient.addColorStop(0.5, 'rgba(255, 0, 0, 0.25)')
    this.gradient.addColorStop(1, 'rgba(255, 0, 0, 0)')

    this.gradient2.addColorStop(0, 'rgba(0, 231, 255, 0.9)')
    this.gradient2.addColorStop(0.5, 'rgba(0, 231, 255, 0.25)')
    this.gradient2.addColorStop(1, 'rgba(0, 231, 255, 0)')

    if (this.color === 1) {
      if (first(this.data.datasets)) {
        first(this.data.datasets).backgroundColor = this.gradient
      }
    } else if (this.color === 2) {
      if (first(this.data.datasets)) {
        first(this.data.datasets).backgroundColor = this.gradient2
      }
    }
    this.renderChart(this.data, this.options)
  }
}
</script>
