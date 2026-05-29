import sys
import os
import numpy as np
from openwakeword.model import Model

# Força o stdout a ser unbuffered (line-buffered) para sincronização em tempo real com o Go
sys.stdout.reconfigure(line_buffering=True)

if len(sys.argv) < 2:
    print("ERROR: Caminho do modelo de wake word nao fornecido.", file=sys.stderr)
    sys.exit(1)

model_path = sys.argv[1]
threshold = 0.8

if len(sys.argv) >= 3:
    try:
        threshold = float(sys.argv[2])
    except ValueError:
        pass

# Verifica se o arquivo do modelo existe
if not os.path.exists(model_path):
    print(f"ERROR: Arquivo do modelo nao encontrado: {model_path}", file=sys.stderr)
    sys.exit(1)

try:
    # Desativa log de warnings desnecessários do openwakeword/onnxruntime
    import warnings
    warnings.filterwarnings("ignore")
    
    # Inicializa o modelo real
    oww_model = Model(wakeword_models=[model_path])
    model_key = list(oww_model.models.keys())[0]
    print(f"READY:{model_key}", flush=True)
except Exception as e:
    print(f"ERROR: Falha ao carregar o modelo openWakeWord: {str(e)}", file=sys.stderr)
    sys.exit(1)

# Bloco de processamento: 1280 amostras de áudio PCM mono de 16-bit a 16kHz
# 1280 amostras * 2 bytes/amostra = 2560 bytes
chunk_size_bytes = 2560

try:
    while True:
        # Lê bytes brutos da entrada padrão (stdin)
        raw_bytes = sys.stdin.buffer.read(chunk_size_bytes)
        if not raw_bytes or len(raw_bytes) < chunk_size_bytes:
            break
        
        # Converte bytes brutos para array numpy int16
        samples = np.frombuffer(raw_bytes, dtype=np.int16)
        
        # Roda inferência de deep learning no openwakeword
        predictions = oww_model.predict(samples)
        
        # Obtém o score de confiança para a palavra-chave
        score = predictions.get(model_key, 0.0)
        print(f"Threshold:{score:.4f}", flush=True)
        # Se ultrapassar o limiar, envia sinal de detecção para o Go
        if score > threshold:
            print(f"DETECTED:{score:.4f}", flush=True)

except KeyboardInterrupt:
    pass
except Exception as e:
    print(f"ERROR: Falha no loop de execucao: {str(e)}", file=sys.stderr)
    sys.exit(1)
