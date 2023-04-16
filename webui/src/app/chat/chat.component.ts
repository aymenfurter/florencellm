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
  isLoading = false;

  sampleQueries = [
    "I am looking for someone that knows the Azure CNI Overlay within Azure Kubernetes service",
    "Who is the right person to talk to when I am interested regarding Azure Container Service workload profiles?",
    "I need help with the Capture cost requirements on Azure. Who is the expert on this?",
    "Who can help me to learn more about Azure Container Instances with VNET integration"
  ];

  constructor(private chatService: ChatService) {}

  async submitMessage(): Promise<void> {
    if (this.userMessage.trim()) {
      this.isLoading = true;

      const data = await this.chatService.sendMessage(this.userMessage, this.messages);

      this.messages = data.messages;
      this.userMessage = '';

      this.isLoading = false;
    }
  }

  // Add the startConversation method
  startConversation(query: string): void {
    this.userMessage = query;
    this.submitMessage();
  }
}
