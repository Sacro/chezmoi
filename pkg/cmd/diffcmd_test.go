	"github.com/alecthomas/assert/v2"
					assert.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				assert.NoError(t, newTestConfig(t, fileSystem, withStdout(&stdout)).execute(append([]string{"diff"}, tc.args...)))