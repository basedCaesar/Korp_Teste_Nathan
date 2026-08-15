export interface ErrorEnvelope {
  code: string;
  message: string;
  details: string[];
  trace_id: string;
}
