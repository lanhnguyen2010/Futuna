const { createApp } = Vue;

createApp({
  data() {
    return {
      date: new Date().toISOString().slice(0,10),
      search: '',
      table: null,
    };
  },
  mounted() {
    this.load();
  },
  watch: {
    date() {
      this.load();
    },
    search() {
      this.filterTable();
    }
  },
  methods: {
    async fetchData() {
      const res = await fetch(`/api/analysis?date=${this.date}`);
      return await res.json();
    },
    async load() {
      const data = await this.fetchData();
      if (this.table) {
        this.table.replaceData(data);
      } else {
        this.table = new Tabulator(this.$refs.table, {
          data,
          layout: "fitColumns",
          columns: [
            { title: "Ticker", field: "ticker" },
            { title: "Short-Term", field: "short_term" },
            { title: "Long-Term", field: "long_term" },
            {
              title: "Overall",
              field: "overall",
              formatter: function(cell) {
                const v = cell.getValue().toLowerCase();
                const color = v.includes('buy') ? 'green' : v.includes('sell') ? 'red' : 'gray';
                cell.getElement().style.color = color;
                return cell.getValue();
              }
            }
          ]
        });
      }
      this.filterTable();
    },
    filterTable() {
      if (!this.table) return;
      const term = this.search.toLowerCase();
      if (term) {
        this.table.setFilter("ticker", "like", term);
      } else {
        this.table.clearFilter();
      }
    }
  }
}).mount('#app');
