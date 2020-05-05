<template>
    <mu-data-table :data="data" :columns="columns">
        <template #default="{row}">
            <td>{{row.Path}}</td>
            <td>{{unitFormat(row.Size)}}</td>
            <td>{{toDurationStr(row.Duration)}}</td>
            <td>
                <mu-button @click="play(row)" icon small>
                    <mu-icon value="play_arrow" />
                </mu-button>
                <mu-button @click="deleteFlv(row)" icon small>
                    <mu-icon value="delete_forever" />
                </mu-button>
            </td>
        </template>
    </mu-data-table>
</template>

<script>
export default {
    data() {
        return {
            data: [],
            columns:[
                {
                    title:"文件路径"
                },{
                    title:"大小"
                },{
                    title:"时长"
                },{title:"操作"}
            ]
        };
    },
    methods: {
        play(item) {
            this.ajax.get(
                "/record/flv/play",
                { streamPath: item.Path.replace(".flv", "") },
                x => {
                    if (x == "success") {
                        this.ajax.getJSON("/record/flv/list", x => this.data = x);
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
        }
    },
    mounted() {
        this.ajax.getJSON("/record/flv/list", x => this.data = x);
    }
};
</script>

