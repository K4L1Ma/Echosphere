package v1

import "go.uber.org/zap/zapcore"

func (a *Ack) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("From", a.GetFrom())
	enc.AddString("To", a.GetTo())
	enc.AddString("Content", a.GetContent())

	return nil
}

func (m *Message) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("From", m.GetFrom())
	enc.AddString("Content", m.GetContent())

	return nil
}

func (r *EchoSphereTransmissionServiceTransmitRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if message := r.GetMessage(); message != nil {
		if err := enc.AddObject("Content", message); err != nil {
			return err
		}
	}

	if ack := r.GetAck(); ack != nil {
		if err := enc.AddObject("Ack", ack); err != nil {
			return err
		}
	}

	return nil
}

func (r *EchoSphereTransmissionServiceTransmitResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if message := r.GetMessage(); message != nil {
		if err := enc.AddObject("Content", message); err != nil {
			return err
		}
	}

	if ack := r.GetAck(); ack != nil {
		if err := enc.AddObject("Ack", ack); err != nil {
			return err
		}
	}

	return nil
}
