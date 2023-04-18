import { Injectable } from '@angular/core';
import { environment  } from 'src/environments/environment';
const API_URL = environment.targetUrl + 'api/conversation'; 

@Injectable({
  providedIn: 'root',
})
export class ChatService {
  constructor() {}

  async sendMessage(
    userMessage: string,
    messages: { role: string; content: string }[]
  ): Promise<{ messages: { role: string; content: string }[]; response: string }> {
    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ userMessage, messages }),
      });

      if (!response.ok) {
        throw new Error(`API request failed: ${response.statusText}`);
      }

      const data = await response.json();
      return data; // Update this line to return the whole data object
    } catch (error) {
      console.error(`Error while sending message: ${(error as Error).message}`);
      return {
        messages: [],
        response: 'Error while sending message',
      };
    }
  }
}
