import { Component } from '@angular/core';
import { ChatService } from '../chat.service';

@Component({
  selector: 'app-chat',
  templateUrl: './chat.component.html',
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent {
  userMessage = '';
  messages: { role: string; content: string }[] = [];

  constructor(private chatService: ChatService) {}

  async submitMessage(): Promise<void> {
    if (this.userMessage.trim()) {
      const data = await this.chatService.sendMessage(this.userMessage, this.messages);
      this.messages = data.messages;
      this.userMessage = '';
    }
  }
  
  
}
