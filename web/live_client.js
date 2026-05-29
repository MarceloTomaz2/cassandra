export class GeminiLiveClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.audioContext = null;
        this.processor = null;
        this.onAudio = null;
        this.onText = null;
        this.onFace = null;
        this.onStatus = null;
        this.onWakeWord = null;
        this.onVolumeChange = null;
        this.isPlaying = false;
        this.audioQueue = [];
        this.nextStartTime = 0;
    }

    async start() {
        this.ws = new WebSocket(this.url);
        this.ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            if (msg.type === 'audio') {
                this.handleAudioResponse(msg.data);
            } else if (msg.type === 'text') {
                if (this.onText) this.onText(msg.data);
            } else if (msg.type === 'status') {
                if (this.onStatus) this.onStatus(msg.data);
            } else if (msg.type === 'wake_word_detected') {
                if (this.onWakeWord) this.onWakeWord(msg.data);
            }
        };

        this.ws.onopen = () => {
            if (this.onStatus) this.onStatus("Conectado");
            this.startRecording();
        };

        this.ws.onclose = () => {
            if (this.onStatus) this.onStatus("Desconectado");
            this.stopRecording();
        };

        this.audioContext = new (window.AudioContext || window.webkitAudioContext)({ sampleRate: 16000 });
        this.outAudioContext = new (window.AudioContext || window.webkitAudioContext)({ sampleRate: 24000 });
    }

    async startRecording() {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        const source = this.audioContext.createMediaStreamSource(stream);
        this.processor = this.audioContext.createScriptProcessor(4096, 1, 1);

        this.processor.onaudioprocess = (e) => {
            const inputData = e.inputBuffer.getChannelData(0);
            
            // Calcula o volume RMS (Root Mean Square) em tempo real
            let sum = 0;
            for (let i = 0; i < inputData.length; i++) {
                sum += inputData[i] * inputData[i];
            }
            let rms = Math.sqrt(sum / inputData.length);
            let volume = Math.min(100, Math.round(rms * 100 * 5)); // Ganho para boa visibilidade
            if (this.onVolumeChange) {
                this.onVolumeChange(volume);
            }

            // Convert to 16-bit PCM
            const pcmData = new Int16Array(inputData.length);
            for (let i = 0; i < inputData.length; i++) {
                pcmData[i] = Math.max(-1, Math.min(1, inputData[i])) * 0x7FFF;
            }
            
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                const base64Audio = btoa(String.fromCharCode(...new Uint8Array(pcmData.buffer)));
                this.ws.send(JSON.stringify({ type: 'audio', data: base64Audio }));
            }
        };

        source.connect(this.processor);
        this.processor.connect(this.audioContext.destination);
    }

    stopRecording() {
        if (this.processor) {
            this.processor.disconnect();
            this.processor = null;
        }
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    handleAudioResponse(base64Data) {
        const binaryString = atob(base64Data);
        const len = binaryString.length;
        const bytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }
        const pcmData = new Int16Array(bytes.buffer);
        
        // Convert PCM to Float32 for AudioContext
        const floatData = new Float32Array(pcmData.length);
        for (let i = 0; i < pcmData.length; i++) {
            floatData[i] = pcmData[i] / 0x7FFF;
        }

        const audioBuffer = this.outAudioContext.createBuffer(1, floatData.length, 24000);
        audioBuffer.getChannelData(0).set(floatData);

        const source = this.outAudioContext.createBufferSource();
        source.buffer = audioBuffer;
        source.connect(this.outAudioContext.destination);

        const currentTime = this.outAudioContext.currentTime;
        if (this.nextStartTime < currentTime) {
            this.nextStartTime = currentTime;
        }
        source.start(this.nextStartTime);
        this.nextStartTime += audioBuffer.duration;

        if (this.onAudio) this.onAudio();
    }
}
