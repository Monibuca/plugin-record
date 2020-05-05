<template>
    <div>
        <Records ref="recordsPanel" v-if="$parent.titleTabActive==1" />
        <stream-table v-else>
            <template v-slot="{row:stream}">
                <m-button v-if="isRecording(stream)" @click="stopRecord(stream)" blink>正在录制</m-button>
                <m-button v-else @click="record(stream)">录制</m-button>
            </template>
        </stream-table>
    </div>
</template>

<script>
import Records from "./components/Records";
export default {
    components: {
        Records
    },
    data() {
        return {};
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
    },
    mounted() {
        this.$parent.titleTabs = ["StreamList", "录制的视频"];
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