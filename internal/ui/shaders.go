package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const blurFS = `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
out vec4 finalColor;
uniform sampler2D texture0;
uniform vec2 resolution;
uniform vec2 direction;
void main() {
    vec2 texel = direction / resolution;
    vec4 c = texture(texture0, fragTexCoord) * 0.2270270270;
    c += texture(texture0, fragTexCoord + texel * 1.3846153846) * 0.3162162162;
    c += texture(texture0, fragTexCoord - texel * 1.3846153846) * 0.3162162162;
    c += texture(texture0, fragTexCoord + texel * 3.2307692308) * 0.0702702703;
    c += texture(texture0, fragTexCoord - texel * 3.2307692308) * 0.0702702703;
    finalColor = c * fragColor;
}
`

var BlurShader rl.Shader
var BlurReady bool
var blurResLoc int32
var blurDirLoc int32

func LoadShaders() {
	BlurShader = rl.LoadShaderFromMemory("", blurFS)
	if BlurShader.ID != 0 {
		BlurReady = true
		blurResLoc = rl.GetShaderLocation(BlurShader, "resolution")
		blurDirLoc = rl.GetShaderLocation(BlurShader, "direction")
	}
}

func UnloadShaders() {
	if BlurReady {
		rl.UnloadShader(BlurShader)
	}
}

func SetBlurResolution(w, h float32) {
	rl.SetShaderValue(BlurShader, blurResLoc, []float32{w, h}, rl.ShaderUniformVec2)
}
func SetBlurDirection(x, y float32) {
	rl.SetShaderValue(BlurShader, blurDirLoc, []float32{x, y}, rl.ShaderUniformVec2)
}
