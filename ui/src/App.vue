<template>
    <div>
        <mu-tabs :value.sync="active1" indicator-color="#80deea" inverse center>
            <mu-tab>直播流</mu-tab>
            <mu-tab>录制的视频</mu-tab>
        </mu-tabs>
        <div v-if="Rooms.length==0 && active1==0" class="empty">
            <Icon type="md-wine" size="50" />没有任何房间
        </div>
        <template v-else-if="active1==0">
            <mu-card v-for="item in Rooms" :key="item.StreamPath" class="room">
                <mu-card-title :title="item.StreamPath" :sub-title="item.StartTime" />
                <mu-card-text>
                    <p>
                        {{SoundFormat(item.AudioInfo.SoundFormat)}} {{item.AudioInfo.PacketCount}}
                        {{SoundRate(item.AudioInfo.SoundRate)}} 声道:{{item.AudioInfo.SoundType}}
                    </p>
                    <p>
                        {{CodecID(item.VideoInfo.CodecID)}} {{item.VideoInfo.PacketCount}}
                        {{item.VideoInfo.SPSInfo.Width}}x{{item.VideoInfo.SPSInfo.Height}}
                    </p>
                </mu-card-text>
                <mu-card-actions>
                    <mu-button icon @click="stopRecord(item)" class="recording" v-if="isRecording(item)">
                        <mu-icon value="fiber_manual_record" />
                    </mu-button>
                    <mu-button icon @click="record(item)" v-else>
                        <mu-icon value="fiber_manual_record" />
                    </mu-button>
                </mu-card-actions>
            </mu-card>
        </template>
        <Records ref="recordsPanel" v-if="active1==1" />
    </div>
</template>

<script>
let roomsES = null;
const SoundFormat = {
    0: "Linear PCM, platform endian",
    1: "ADPCM",
    2: "MP3",
    3: "Linear PCM, little endian",
    4: "Nellymoser 16kHz mono",
    5: "Nellymoser 8kHz mono",
    6: "Nellymoser",
    7: "G.711 A-law logarithmic PCM",
    8: "G.711 mu-law logarithmic PCM",
    9: "reserved",
    10: "AAC",
    11: "Speex",
    14: "MP3 8Khz",
    15: "Device-specific sound"
};
const CodecID = {
    1: "JPEG (currently unused)",
    2: "Sorenson H.263",
    3: "Screen video",
    4: "On2 VP6",
    5: "On2 VP6 with alpha channel",
    6: "Screen video version 2",
    7: "AVC",
    12: "H265"
};
import Records from "./components/Records";
export default {
    components: {
        Records
    },
    data() {
        return {
            Rooms: [],
            typeMap: {
                Receiver: "📡",
                FlvFile: "🎥",
                TS: "🎬",
                HLS: "🍎",
                "": "⏳",
                Match365: "🏆",
                RTMP: "🚠"
            }
        };
    },
    methods: {
        SoundFormat(soundFormat) {
            return SoundFormat[soundFormat];
        },
        CodecID(codec) {
            return CodecID[codec];
        },
        SoundRate(rate) {
            return rate > 1000 ? rate / 1000 + "kHz" : rate + "Hz";
        },
        record(item) {
            this.$Modal.confirm({
                title: "提示",
                content:
                    "<p>是否使用追加模式</p><small>选择取消将覆盖已有文件</small>",
                onOk: () => {
                    window.ajax.get(
                        "/record/flv?append=true",
                        { streamPath: item.StreamPath },
                        x => {
                            if (x == "success") {
                                this.$Message.success("开始录制(追加模式)");
                            } else {
                                this.$Message.error(x);
                            }
                        }
                    );
                },
                onCancel: () => {
                    window.ajax.get(
                        "/record/flv",
                        { streamPath: item.StreamPath },
                        x => {
                            if (x == "success") {
                                this.$Message.success("开始录制");
                            } else {
                                this.$Message.error(x);
                            }
                        }
                    );
                }
            });
        },
        stopRecord(item) {
            window.ajax.get(
                "/record/flv/stop",
                { streamPath: item.StreamPath },
                x => {
                    if (x == "success") {
                        this.$Message.success("停止录制");
                    } else {
                        this.$Message.error(x);
                    }
                }
            );
        },
        isRecording(item) {
            return (
                item.SubscriberInfo &&
                item.SubscriberInfo.find(x => x.Type == "FlvRecord")
            );
        },
        fetchRooms() {
            roomsES = new EventSource("/api/summary");
            roomsES.onmessage = evt => {
                if (!evt.data) return;
                let summary = JSON.parse(evt.data);
                this.Rooms = (summary && summary.Rooms) || [];
                this.Rooms.sort((a, b) =>
                    a.StreamPath > b.StreamPath ? 1 : -1
                );
            };
        },
        onClickTab(name) {
            this.$refs.recordsPanel.onVisible(name == "recordsPanel");
        }
    },
    mounted() {
        this.fetchRooms();
    },
    destroyed() {
        roomsES.close();
    }
};
</script>

<style>
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