'use client';

import { Hash, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';

// Slack Aubergine palette
const SLACK = {
  sidebar: '#3f0e40',
  sidebarHover: '#350d36',
  sidebarActive: '#1164a3',
  channelHeader: '#350d36',
  channelHeaderBorder: 'rgba(255,255,255,0.1)',
  bg: '#ffffff',
  messageMeta: '#616061',
  messageText: '#1d1c1d',
  inputBg: '#f8f8f8',
  inputBorder: '#e8e8e8',
  inputPlaceholder: '#616061',
  avatarYou: '#611f69',
  avatarOther: '#1264a3',
  accent: '#611f69',
} as const;

const slackMessages = [
  {
    user: 'Alex',
    avatar: 'A',
    time: '10:24 AM',
    message: 'Hey, can someone share the .env file for the staging env?',
    isUser: false,
  },
  {
    user: 'Sarah',
    avatar: 'S',
    time: '10:25 AM',
    message: "I'll DM you the DATABASE_URL",
    isUser: false,
  },
  {
    user: 'Alex',
    avatar: 'A',
    time: '10:25 AM',
    message: 'Thanks! Also need the API keys for the payment service',
    isUser: false,
  },
  {
    user: 'You',
    avatar: 'Y',
    time: '10:26 AM',
    message: 'Stop sharing secrets in Slack. Use openenvx instead.',
    isUser: true,
  },
];

const channels = [
  { name: 'general', unread: true },
  { name: 'random', unread: false },
  { name: 'dev-secrets', unread: false },
];

function TypingIndicator() {
  return (
    <div aria-hidden className="flex items-center gap-1 py-1">
      <span
        className="size-1.5 animate-bounce rounded-full bg-[#616061]"
        style={{ animationDelay: '0ms' }}
      />
      <span
        className="size-1.5 animate-bounce rounded-full bg-[#616061]"
        style={{ animationDelay: '150ms' }}
      />
      <span
        className="size-1.5 animate-bounce rounded-full bg-[#616061]"
        style={{ animationDelay: '300ms' }}
      />
    </div>
  );
}

export function SlackMock() {
  const [visibleMessages, setVisibleMessages] = useState<number[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [typingUser, setTypingUser] = useState<string | null>(null);

  useEffect(() => {
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    const delays = [0, 2000, 4500, 7000, 10_000];

    const runAnimation = () => {
      setVisibleMessages([]);
      setIsTyping(false);
      setTypingUser(null);

      delays.forEach((delay, index) => {
        const showTypingTimeout = setTimeout(() => {
          if (index < slackMessages.length) {
            setTypingUser(slackMessages[index].user);
            setIsTyping(true);
          }
        }, delay);

        const showMessageTimeout = setTimeout(() => {
          setIsTyping(false);
          setTypingUser(null);
          setVisibleMessages((prev) => [...prev, index]);
        }, delay + 800);

        timeouts.push(showTypingTimeout, showMessageTimeout);
      });

      const restartTimeout = setTimeout(() => {
        runAnimation();
      }, 16_000);

      timeouts.push(restartTimeout);
    };

    runAnimation();

    return () => {
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <div className="flex h-full min-h-[380px] flex-col bg-white md:min-h-[400px]">
      {/* Slack-style layout: sidebar + main */}
      <div className="flex min-h-0 flex-1">
        {/* Left sidebar - channel list (Aubergine) */}
        <div
          className="flex w-[52px] shrink-0 flex-col items-center gap-1 border-r py-2"
          style={{
            backgroundColor: SLACK.sidebar,
            borderColor: SLACK.channelHeaderBorder,
          }}
        >
          <div
            className="flex size-8 items-center justify-center rounded-md"
            style={{ backgroundColor: SLACK.sidebarHover }}
          >
            <Plus aria-hidden className="size-4 text-white opacity-90" />
          </div>
          <div
            className="my-1 h-px w-6 opacity-30"
            style={{ backgroundColor: 'white' }}
          />
          {channels.map((ch, i) => (
            <div
              className={`flex size-8 items-center justify-center rounded-md ${i === 0 ? 'bg-white/20' : ''}`}
              key={ch.name}
              style={i > 0 ? { backgroundColor: 'transparent' } : undefined}
              title={ch.name}
            >
              <Hash aria-hidden className="size-4 text-white opacity-95" />
            </div>
          ))}
        </div>

        {/* Main: channel header + messages + input */}
        <div className="flex min-w-0 flex-1 flex-col bg-white">
          {/* Channel header */}
          <div
            className="flex shrink-0 items-center gap-3 border-b px-4 py-2.5"
            style={{
              backgroundColor: SLACK.channelHeader,
              borderColor: SLACK.channelHeaderBorder,
            }}
          >
            <Hash aria-hidden className="size-4 text-white/90" />
            <span className="font-semibold text-sm text-white">general</span>
            <span className="text-white/60 text-xs">240 members</span>
          </div>

          {/* Messages */}
          <div className="flex-1 overflow-y-auto px-4 py-3">
            <div className="space-y-1">
              {slackMessages.map(
                (msg, i) =>
                  visibleMessages.includes(i) && (
                    <div
                      className="-mx-2 flex animate-[fadeInUp_0.3s_ease-out] gap-3 rounded-lg px-2 py-1.5 transition-colors hover:bg-[#f8f8f8]"
                      key={`${msg.user}-${msg.time}-${i}`}
                    >
                      <div
                        className="flex size-9 shrink-0 items-center justify-center rounded-full font-semibold text-[11px] text-white"
                        style={{
                          backgroundColor: msg.isUser
                            ? SLACK.avatarYou
                            : SLACK.avatarOther,
                        }}
                      >
                        {msg.avatar}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-baseline gap-2">
                          <span
                            className="font-semibold text-[13px]"
                            style={{ color: SLACK.messageText }}
                          >
                            {msg.user}
                          </span>
                          <span
                            className="text-[11px]"
                            style={{ color: SLACK.messageMeta }}
                          >
                            {msg.time}
                          </span>
                        </div>
                        <p
                          className="mt-0.5 text-[13px] leading-[1.4]"
                          style={{
                            color: msg.isUser
                              ? SLACK.accent
                              : SLACK.messageText,
                            fontWeight: msg.isUser ? 600 : 400,
                          }}
                        >
                          {msg.message}
                        </p>
                      </div>
                    </div>
                  )
              )}
              {isTyping && typingUser && (
                <div className="-mx-2 flex animate-[fadeInUp_0.2s_ease-out] gap-3 rounded-lg px-2 py-1.5">
                  <div
                    className="flex size-9 shrink-0 items-center justify-center rounded-full font-semibold text-[11px] text-white"
                    style={{ backgroundColor: SLACK.avatarOther }}
                  >
                    {slackMessages.find((m) => m.user === typingUser)?.avatar ??
                      typingUser[0]}
                  </div>
                  <div className="flex min-w-0 flex-1 items-center gap-2 pt-0.5">
                    <span
                      className="font-semibold text-[13px]"
                      style={{ color: SLACK.messageText }}
                    >
                      {typingUser}
                    </span>
                    <TypingIndicator />
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Message input */}
          <div
            className="shrink-0 border-t px-4 py-3"
            style={{
              backgroundColor: SLACK.bg,
              borderColor: SLACK.inputBorder,
            }}
          >
            <div
              className="rounded-lg border px-3 py-2"
              style={{
                backgroundColor: SLACK.inputBg,
                borderColor: SLACK.inputBorder,
              }}
            >
              <span
                className="text-[13px]"
                style={{ color: SLACK.inputPlaceholder }}
              >
                Message #general
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
