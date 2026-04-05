import { describe, expect, it } from "vitest";
import { normalizeTopic, parseDeviceRowIdFromTopic } from "./topic.js";

describe("normalizeTopic", () => {
  it("prefixes relative topics", () => {
    expect(normalizeTopic("lab/d1/cmd", "flora")).toBe("flora/lab/d1/cmd");
  });
  it("leaves already-prefixed topics", () => {
    expect(normalizeTopic("flora/lab/d1/cmd", "flora")).toBe("flora/lab/d1/cmd");
  });
  it("throws on empty", () => {
    expect(() => normalizeTopic("  ", "flora")).toThrow("topic is empty");
  });
});

describe("parseDeviceRowIdFromTopic", () => {
  it("reads device row id as first segment after prefix", () => {
    expect(
      parseDeviceRowIdFromTopic(
        "flora/550e8400-e29b-41d4-a716-446655440000/heartbeat",
        "flora"
      )
    ).toBe("550e8400-e29b-41d4-a716-446655440000");
  });
});
