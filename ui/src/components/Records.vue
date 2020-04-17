<template>
    <div class="records">
        <mu-card v-for="item in data" :key="item">
            <mu-card-title :title="item.Path" :sub-title="unitFormat(item.Size)+' '+toDurationStr(item.Duration)" />
            <mu-card-actions>
                <mu-button @click="play(item)" icon small>
                    <mu-icon value="play_arrow" />
                </mu-button>
                <mu-button @click="deleteFlv(item)" icon small>
                    <mu-icon value="delete_forever" />
                </mu-button>
            </mu-card-actions>
        </mu-card>
    </div>
</template>

<script>
export default {
    data() {
        return {
            data: []
        };
    },
    methods: {
        play(item) {
            this.ajax.get(
                "/record/flv/play",
                { streamPath: item.Path.replace(".flv", "") },
                x => {
                    if (x == "success") {
                        this.onVisible(true);
                        this.$toast.success("开始发布");
                    } else {
                        this.$toast.error(x);
                    }
                }
            );
        },
        deleteFlv(item) {
            this.$confirm("是否删除Flv文件", "提示").then(result => {
                if (result) {
                    return this.ajax.get(
                        "/record/flv/delete",
                        { streamPath: item.Path.replace(".flv", "") },
                        x => {
                            if (x == "success") {
                                this.$toast.success("删除成功");
                            } else {
                                this.$toast.error(x);
                            }
                        }
                    );
                }
            });
        },
        toDurationStr(value) {
            if (value > 1000) {
                let s = value / 1000;
                if (s > 60) {
                    s = s | 0;
                    let min = (s / 60) >> 0;
                    if (min > 60) {
                        let hour = (min / 60) >> 0;
                        return hour + "hour" + (min % 60) + "min";
                    } else {
                        return min + "min" + (s % 60) + "s";
                    }
                } else {
                    return s.toFixed(3) + "s";
                }
            } else {
                return value + "ms";
            }
        },
        onVisible(visible) {
            if (visible) {
                this.ajax.getJSON("/record/flv/list", {}, x => {
                    this.data = x;
                });
            }
        }
    }
};
</script>

<style scoped>
.records {
    display: flex;
    flex-wrap: wrap;
    padding: 0 15px;
}
.records > * {
    width: 200px;
}
</style>