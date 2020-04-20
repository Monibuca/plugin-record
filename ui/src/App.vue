<template>
    <div>
        <mu-tabs :value.sync="active1" indicator-color="#80deea" inverse center>
            <mu-tab>直播流</mu-tab>
            <mu-tab>录制的视频</mu-tab>
        </mu-tabs>
        <mu-data-table v-if="active1==0" :columns="columns" :data="$store.state.Room" :min-col-width="50"
            @row-click="(i,r)=>isRecording(r)?stopRecord(r):record(r)">
            <template slot-scope="scope">
                <td class="is-center" v-if="isRecording(scope.row)"></td>
                <td class="is-center" v-else></td>
                <td class="is-center">{{scope.row.StreamPath}}</td>
                <td class="is-center">{{scope.row.Type||"await"}}</td>
                <td class="is-center">
                    <StartTime :value="scope.row.StartTime"></StartTime>
                </td>
                <td class="is-center">{{SoundFormat(scope.row.AudioInfo.SoundFormat)}}</td>
                <td class="is-center">{{SoundRate(scope.row.AudioInfo.SoundRate)}}</td>
                <td class="is-center">{{scope.row.AudioInfo.SoundType}}</td>
                <td class="is-center">{{CodecID(scope.row.VideoInfo.CodecID)}}</td>
                <td class="is-center">{{scope.row.VideoInfo.SPSInfo.Width}}x{{scope.row.VideoInfo.SPSInfo.Height}}</td>
                <td class="is-center">{{scope.row.AudioInfo.PacketCount}}/{{scope.row.VideoInfo.PacketCount}}</td>
                <td class="is-center">{{getSubscriberCount(scope.row)}}</td>
            </template>
        </mu-data-table>
        <Records ref="recordsPanel" v-if="active1==1" />
    </div>
</template>

<script>
import Records from "./components/Records";
export default {
    components: {
        Records
    },
    data() {
        return {
            columns: [
                {
                    title: "房间",
                    name: "StreamPath",
                    sortable: true
                },
                {
                    title: "类型",
                    name: "Type",
                    sortable: true
                },
                {
                    title: "开始时间",
                    name: "StartTime",
                    sortable: true
                },
                {
                    title: "音频格式",
                    name: "AudioInfo"
                },
                {
                    title: "采样率",
                    name: "AudioInfo"
                },
                {
                    title: "声道",
                    name: "AudioInfo"
                },
                {
                    title: "视频格式",
                    name: "VideoInfo"
                },
                {
                    title: "分辨率",
                    name: "VideoInfo"
                },
                {
                    title: "数据包",
                    name: ""
                },
                {
                    title: "订阅者",
                    name: "Subscribes"
                }
            ]
        };
    },
    methods: {
        record(item) {
            let append = false;
            this.$confirm(
                h =>
                    h("mu-switch", {
                        props: {
                            label: "追加模式"
                        },
                        on: {
                            change(value) {
                                append = value;
                            }
                        }
                    }),
                "是否开始录制"
            ).then(result => {
                if (result) {
                    this.ajax.get(
                        "/record/flv?append=" + append,
                        { streamPath: item.StreamPath },
                        x => {
                            if (x == "success") {
                                this.$toast.success(
                                    "开始录制" + (append ? "(追加模式)" : "")
                                );
                            } else {
                                this.$toast.error(x);
                            }
                        }
                    );
                }
            });
        },
        stopRecord(item) {
            this.$confirm("是否停止录制", "提示").then(result => {
                this.ajax.get(
                    "/record/flv/stop",
                    { streamPath: item.StreamPath },
                    x => {
                        if (x == "success") {
                            this.$toast.success("停止录制");
                        } else {
                            this.$toast.error(x);
                        }
                    }
                );
            });
        },
        isRecording(item) {
            return (
                item.SubscriberInfo &&
                item.SubscriberInfo.find(x => x.Type == "FlvRecord")
            );
        }
    }
};
</script>

<style scoped>
@keyframes recording {
    0% {
        opacity: 0.2;
    }
    50% {
        opacity: 1;
    }
    100% {
        opacity: 0.2;
    }
}

.recording {
    animation: recording 1s infinite;
}

.layout {
    padding-bottom: 30px;
    display: flex;
    flex-wrap: wrap;
}

.room {
    width: 250px;
    margin: 10px;
    text-align: left;
}

.empty {
    color: #ffc107;
    width: 100%;
    min-height: 500px;
    display: flex;
    justify-content: center;
    align-items: center;
}

.status {
    position: fixed;
    display: flex;
    left: 5px;
    bottom: 10px;
}

.status > div {
    margin: 0 5px;
}
</style>