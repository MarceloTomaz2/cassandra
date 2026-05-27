export async function talkToBot(userInput) {
    try {
        const response = await fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message: userInput })
        });

        const data = await response.json();
        if (data.response) {
            return data.response;
        } else {
            return "Minha conexão com o servidor caiu. Deve ser interferência magnética.";
        }
    } catch (error) {
        console.error("Erro na chamada Proxy:", error);
        return "Erro de rede. Alguém deve ter desligado o roteador.";
    }
}
