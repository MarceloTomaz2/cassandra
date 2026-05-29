import * as THREE from 'three';
import { GeminiLiveClient } from './live_client.js';

// Configuração básica
const canvas = document.querySelector('#scene');
const renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
renderer.shadowMap.enabled = true;
renderer.outputColorSpace = THREE.SRGBColorSpace;

const scene = new THREE.Scene();
scene.background = new THREE.Color(0x000000);

const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);

const IntensidadePerspectiva = 0.5; 

// CONFIGURAÇÃO DINÂMICA DO QUARTO
let roomGroup = new THREE.Group();
scene.add(roomGroup);

function createTiledFace(width, height, rotation, position) {
    const group = new THREE.Group();
    const gridSize = 10;
    const tileW = width / gridSize, tileH = height / gridSize;
    for (let i = 0; i < gridSize; i++) {
        for (let j = 0; j < gridSize; j++) {
            const color = new THREE.Color().setHSL(Math.random(), 0.7, 0.4);
            const tile = new THREE.Mesh(new THREE.PlaneGeometry(tileW, tileH), new THREE.MeshStandardMaterial({ color, roughness: 0.2, metalness: 0.1 }));
            tile.receiveShadow = true;
            tile.position.set((i * tileW) - (width/2) + (tileW/2), (j * tileH) - (height/2) + (tileH/2), 0);
            group.add(tile);
        }
    }
    group.rotation.copy(rotation);
    group.position.copy(position);
    return group;
}

function updateRoomDimensions() {
    scene.remove(roomGroup);
    roomGroup = new THREE.Group();
    
    const roomW = Math.max(3, window.innerWidth / 200);
    const roomH = Math.max(3, window.innerHeight / 200);
    const roomD = 5; 
    
    const zCenter = 0.45;

    roomGroup.add(createTiledFace(roomW, roomD, new THREE.Euler(-Math.PI/2, 0, 0), new THREE.Vector3(0, 0, zCenter)));
    roomGroup.add(createTiledFace(roomW, roomD, new THREE.Euler(Math.PI/2, 0, 0), new THREE.Vector3(0, roomH, zCenter)));
    roomGroup.add(createTiledFace(roomD, roomH, new THREE.Euler(0, Math.PI/2, 0), new THREE.Vector3(-roomW/2, roomH/2, zCenter)));
    roomGroup.add(createTiledFace(roomD, roomH, new THREE.Euler(0, -Math.PI/2, 0), new THREE.Vector3(roomW/2, roomH/2, zCenter)));
    roomGroup.add(createTiledFace(roomW, roomH, new THREE.Euler(0, 0, 0), new THREE.Vector3(0, roomH/2, zCenter - roomD/2)));
    roomGroup.add(createTiledFace(roomW, roomH, new THREE.Euler(0, Math.PI, 0), new THREE.Vector3(0, roomH/2, zCenter + roomD/2)));
    
    scene.add(roomGroup);
}

updateRoomDimensions();

// ROBÔ (BENDER)
const robot = new THREE.Group();
const metalMat = new THREE.MeshStandardMaterial({ color: 0xdddddd, metalness: 0.8, roughness: 0.2 });
const legL = new THREE.Mesh(new THREE.CylinderGeometry(0.1, 0.1, 0.8), metalMat); legL.position.set(-0.25, 0.4, 0); robot.add(legL);
const legR = new THREE.Mesh(new THREE.CylinderGeometry(0.1, 0.1, 0.8), metalMat); legR.position.set(0.25, 0.4, 0); robot.add(legR);
const torso = new THREE.Mesh(new THREE.CylinderGeometry(0.35, 0.35, 0.6), metalMat); torso.position.y = 1.1; robot.add(torso);
const headMain = new THREE.Mesh(new THREE.CylinderGeometry(0.2, 0.2, 0.3), metalMat); headMain.position.y = 1.55; robot.add(headMain);
const dome = new THREE.Mesh(new THREE.SphereGeometry(0.2, 16, 16, 0, Math.PI*2, 0, Math.PI/2), metalMat); dome.position.y = 1.7; robot.add(dome);
const visor = new THREE.Mesh(new THREE.BoxGeometry(0.35, 0.15, 0.05), new THREE.MeshBasicMaterial({ color: 0x000000 })); visor.position.set(0, 1.6, 0.18); robot.add(visor);

function createEye(x) {
    const group = new THREE.Group();
    const sclera = new THREE.Mesh(new THREE.PlaneGeometry(0.12, 0.1), new THREE.MeshBasicMaterial({ color: 0xffffff }));
    const iris = new THREE.Mesh(new THREE.PlaneGeometry(0.04, 0.04), new THREE.MeshBasicMaterial({ color: 0x000000 }));
    iris.position.z = 0.01; group.add(sclera, iris); group.position.set(x, 1.6, 0.21);
    return { group, iris, sclera };
}
const eyeL = createEye(-0.08), eyeR = createEye(0.08); robot.add(eyeL.group, eyeR.group);

const mouthWidth = 0.15, mouthSegments = 20;
const mouthGeometry = new THREE.BufferGeometry();
const mouthPoints = [];
for (let i = 0; i <= mouthSegments; i++) mouthPoints.push(new THREE.Vector3((i/mouthSegments)*mouthWidth - mouthWidth/2, 0, 0));
mouthGeometry.setFromPoints(mouthPoints);
const mouthLine = new THREE.Line(mouthGeometry, new THREE.LineBasicMaterial({ color: 0x00ff00 }));
mouthLine.position.set(0, 1.48, 0.21); robot.add(mouthLine);
scene.add(robot);

// CÂMERA E CONTROLES
const eyeLevel = 1.6, POINT_ZERO = 0.9;
let cameraDistance = POINT_ZERO, mouseX = 0, mouseY = 0, targetX = 0, targetY = 0;

window.addEventListener('wheel', (e) => cameraDistance = Math.min(POINT_ZERO, Math.max(0.3, cameraDistance + e.deltaY * 0.002)));

window.addEventListener('resize', () => {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
    updateRoomDimensions();
});

scene.add(new THREE.AmbientLight(0xffffff, 0.5));
const light = new THREE.DirectionalLight(0xffffff, 1.2); light.position.set(2, 5, 5); light.castShadow = true; scene.add(light);

// LÓGICA DE CONVERSA EM TEMPO REAL E EVENTOS
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

// Rastreamento Facial contínuo via /events
const eventsWs = new WebSocket(`${protocol}//${window.location.host}/events`);
eventsWs.onmessage = (event) => {
    try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'face') {
            mouseX = -msg.x; // Invertido para movimento espelhado natural
            mouseY = msg.y;
        }
    } catch (e) {
        console.error("Error parsing event:", e);
    }
};

let isSpeaking = false;
let speakingTimeout;
const micBtn = document.querySelector('#micBtn');
const statusDiv = document.querySelector('#status');

const liveClient = new GeminiLiveClient(`${protocol}//${window.location.host}/ws`);

const volumeMeter = document.querySelector('#volume-meter');
const volumeBar = document.querySelector('#volume-bar');

liveClient.onStatus = (status) => {
    statusDiv.innerText = status;
};

liveClient.onAudio = () => {
    isSpeaking = true;
    clearTimeout(speakingTimeout);
    speakingTimeout = setTimeout(() => {
        isSpeaking = false;
    }, 200);
};

liveClient.onText = (text) => {
    console.log("Gemini:", text);
};

liveClient.onVolumeChange = (volume) => {
    if (volumeBar) {
        volumeBar.style.width = volume + '%';
    }
};

let isActive = false;
micBtn.addEventListener('click', async () => {
    if (!isActive) {
        await liveClient.start();
        micBtn.style.background = "rgba(0, 255, 0, 0.3)";
        micBtn.innerText = "🛑 Desativar Escuta";
        if (volumeMeter) volumeMeter.style.display = 'block';
        isActive = true;
    } else {
        liveClient.stopRecording();
        micBtn.style.background = "rgba(255, 255, 255, 0.2)";
        micBtn.innerText = "🎙️ Ativar Escuta Ativa";
        if (volumeMeter) volumeMeter.style.display = 'none';
        if (volumeBar) volumeBar.style.width = '0%';
        isActive = false;
        isSpeaking = false;
        statusDiv.innerText = "Clique para ativar...";
    }
});

function playNotificationSound(type) {
    try {
        const audioCtx = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = audioCtx.createOscillator();
        const gainNode = audioCtx.createGain();

        oscillator.connect(gainNode);
        gainNode.connect(audioCtx.destination);

        if (type === 'active') {
            oscillator.type = 'sine';
            oscillator.frequency.setValueAtTime(587.33, audioCtx.currentTime); // D5
            gainNode.gain.setValueAtTime(0.08, audioCtx.currentTime);
            oscillator.start();
            oscillator.frequency.setValueAtTime(880, audioCtx.currentTime + 0.08); // A5
            gainNode.gain.exponentialRampToValueAtTime(0.001, audioCtx.currentTime + 0.3);
            oscillator.stop(audioCtx.currentTime + 0.32);
        } else {
            oscillator.type = 'sine';
            oscillator.frequency.setValueAtTime(440, audioCtx.currentTime); // A4
            gainNode.gain.setValueAtTime(0.08, audioCtx.currentTime);
            oscillator.start();
            oscillator.frequency.setValueAtTime(293.66, audioCtx.currentTime + 0.08); // D4
            gainNode.gain.exponentialRampToValueAtTime(0.001, audioCtx.currentTime + 0.3);
            oscillator.stop(audioCtx.currentTime + 0.32);
        }
    } catch (e) {
        console.error("Erro ao reproduzir som de notificação:", e);
    }
}

liveClient.onWakeWord = (state) => {
    if (state === 'active') {
        console.log("Wake word detectada! Bender ativado!");
        micBtn.style.background = "rgba(255, 0, 0, 0.6)"; // Vermelho para modo conversa ativo!
        micBtn.innerText = "🛑 Parar Conversa";
        playNotificationSound('active');
    } else {
        console.log("Bender voltou ao modo de escuta ativa...");
        micBtn.style.background = "rgba(0, 255, 0, 0.3)"; // Verde para escuta ativa aguardando
        micBtn.innerText = "🛑 Desativar Escuta";
        playNotificationSound('idle');
    }
};

let blinkState = 0;
let nextBlinkTime = Date.now() + Math.random() * 10000;
let blinkEndTime = 0;

function animate() {
    requestAnimationFrame(animate);
    const time = Date.now() * 0.01;
    const positions = mouthLine.geometry.attributes.position.array;
    for (let i = 0; i <= mouthSegments; i++) {
        const x = (i / mouthSegments);
        const amplitude = isSpeaking ? 0.06 : 0.005; 
        positions[i * 3 + 1] = Math.sin(x * 25 + time) * amplitude * Math.sin(time * 0.8);
    }
    mouthLine.geometry.attributes.position.needsUpdate = true;

    targetX += (mouseX - targetX) * 0.05;
    targetY += (mouseY - targetY) * 0.05;
    camera.position.x = targetX * (1.5 * IntensidadePerspectiva);
    camera.position.y = eyeLevel + (targetY * (1.5 * IntensidadePerspectiva));
    camera.position.z = cameraDistance;
    camera.rotation.y = targetX * (1.5 * IntensidadePerspectiva); 
    camera.rotation.x = -targetY * (1.0 * IntensidadePerspectiva);

    const irisMax = 0.04;
    eyeL.iris.position.x = THREE.MathUtils.clamp(targetX * 0.05, -irisMax, irisMax);
    eyeL.iris.position.y = THREE.MathUtils.clamp(targetY * 0.04, -irisMax, irisMax);
    eyeR.iris.position.x = eyeL.iris.position.x;
    eyeR.iris.position.y = eyeL.iris.position.y;

    // Fazer o corpo do robô girar levemente na direção da face
    robot.rotation.y = targetX * 0.3;
    robot.rotation.x = -targetY * 0.1;

    // Lógica de piscar os olhos (TV sem sintonia)
    const nowBlink = Date.now();
    if (blinkState === 0 && nowBlink > nextBlinkTime) {
        blinkState = 1;
        blinkEndTime = nowBlink + 150 + Math.random() * 300; // Pisca de 150ms a 450ms
    }
    if (blinkState === 1) {
        if (nowBlink > blinkEndTime) {
            blinkState = 0;
            nextBlinkTime = nowBlink + Math.random() * 10000;
            eyeL.sclera.scale.set(1, 1, 1);
            eyeR.sclera.scale.set(1, 1, 1);
            eyeL.sclera.material.color.setHex(0xffffff);
            eyeR.sclera.material.color.setHex(0xffffff);
            eyeL.iris.visible = true;
            eyeR.iris.visible = true;
        } else {
            // Efeito de TV sem sintonia
            const scaleY = Math.random() * 0.5; // Esmaga os olhos verticalmente
            eyeL.sclera.scale.set(1, scaleY, 1);
            eyeR.sclera.scale.set(1, scaleY, 1);
            
            const gray = Math.random() * 0.5 + 0.5; // Tons de cinza para estática
            eyeL.sclera.material.color.setRGB(gray, gray, gray);
            eyeR.sclera.material.color.setRGB(gray, gray, gray);
            
            const showIris = Math.random() > 0.4; // Falha na íris
            eyeL.iris.visible = showIris;
            eyeR.iris.visible = showIris;
        }
    }

    renderer.render(scene, camera);
}
animate();
