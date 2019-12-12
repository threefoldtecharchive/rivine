<script>
import { Doughnut } from 'vue-chartjs'
import { first } from 'lodash'

export default {
  extends: Doughnut,
  props: ['data', 'options', 'color'],
  mounted() {
    this.renderChart(this.data, this.options);
  }
};
</script>
